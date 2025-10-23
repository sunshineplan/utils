// https://github.com/scorredoira/email
package mail

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"mime"
	"net/mail"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var contentTypes = [...]string{"text/plain", "text/html"}

// ContentType represents content type
type ContentType int

const (
	// TextPlain sets body type to text/plain in message body
	TextPlain ContentType = iota
	// TextHTML sets body type to text/html in message body
	TextHTML
)

func (contentType ContentType) String() string {
	return contentTypes[contentType]
}

// Attachment represents an email attachment
type Attachment struct {
	Filename  string
	Path      string
	Bytes     []byte
	ContentID string
}

// Message represents an email message
type Message struct {
	From        *mail.Address
	To, Cc, Bcc Receipts
	Subject     string
	Body        string
	ContentType ContentType
	Attachments []*Attachment
}

// RcptList returns a de-duplicated list of recipient email addresses while preserving order:
// first To, then Cc, then Bcc.
func (m *Message) RcptList() (rcpts []string) {
	seen := make(map[string]bool)
	for _, list := range [][]*mail.Address{m.To, m.Cc, m.Bcc} {
		for _, addr := range list {
			if !seen[addr.Address] {
				rcpts = append(rcpts, addr.Address)
				seen[addr.Address] = true
			}
		}
	}
	return
}

// Bytes renders the RFC822-style message bytes for the message.
// id is used to create the Message-ID domain if provided; if empty, fallback to hostname.
// The produced message uses CRLF line endings as required by SMTP and includes correct
// MIME headers for single-part or multipart/mixed with attachments.
func (m *Message) Bytes(id string) []byte {
	// determine hostname part for Message-ID
	if id == "" {
		if m.From != nil && m.From.Address != "" {
			id = m.From.Address
		} else {
			if hostname, _ := os.Hostname(); hostname != "" {
				id = hostname
			} else {
				id = "localhost"
			}
		}
	}
	id = generateMsgID(id)

	var buf bytes.Buffer
	w := textproto.NewWriter(bufio.NewWriter(&buf))

	// Basic headers
	w.PrintfLine("MIME-Version: 1.0")
	w.PrintfLine("Date: %s", time.Now().Format(time.RFC1123Z))
	w.PrintfLine("Message-ID: <%s>", id)
	w.PrintfLine("Subject: %s", encodeHeader(m.Subject))
	if m.From != nil {
		w.PrintfLine("From: %s", m.From.String())
	}
	// To / Cc headers (these are header fields visible in the message)
	if len(m.To) > 0 {
		w.PrintfLine("To: %s", m.To)
	}
	if len(m.Cc) > 0 {
		w.PrintfLine("Cc: %s", m.Cc)
	}

	boundary := randomString(16)
	if len(m.Attachments) > 0 {
		w.PrintfLine("Content-Type: multipart/mixed; boundary=%q", boundary)
		w.PrintfLine("")
		w.PrintfLine("--%s", boundary)
	}

	w.PrintfLine(`Content-Type: %s; charset="UTF-8"`, m.ContentType)
	w.PrintfLine("Content-Transfer-Encoding: base64")
	w.PrintfLine("")
	writeBase64BytesLines(w, []byte(m.Body))

	if l := len(m.Attachments); l > 0 {
		for i, attachment := range m.Attachments {
			w.PrintfLine("--%s", boundary)
			if mimetype := mime.TypeByExtension(filepath.Ext(attachment.Filename)); mimetype != "" {
				w.PrintfLine("Content-Type: %s", mimetype)
			} else {
				w.PrintfLine("Content-Type: application/octet-stream")
			}
			if attachment.ContentID != "" {
				w.PrintfLine(`Content-Disposition: inline; filename="%s"`, encodeHeader(attachment.Filename))
				w.PrintfLine("Content-ID: <%s>", attachment.ContentID)
			} else {
				w.PrintfLine(`Content-Disposition: attachment; filename="%s"`, encodeHeader(attachment.Filename))
			}
			w.PrintfLine("Content-Transfer-Encoding: base64")
			w.PrintfLine("")
			writeBase64BytesLines(w, attachment.Bytes)

			if i < l-1 {
				w.PrintfLine("--%s", boundary)
			} else {
				w.PrintfLine("--%s--", boundary)
			}
		}
	}

	return buf.Bytes()
}

// encodeHeader encodes a header value using RFC2047 only when non-ASCII chars are present.
// For pure ASCII strings it returns the original string unmodified.
func encodeHeader(s string) string {
	for _, r := range s {
		if r > 127 {
			return fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(s)))
		}
	}
	return s
}

// writeBase64BytesLines encodes bytes to base64 and writes it in lines of up to 76 chars.
func writeBase64BytesLines(w *textproto.Writer, b []byte) {
	enc := base64.StdEncoding.EncodeToString(b)
	// number of lines
	for i := 0; i < len(enc); i += 76 {
		end := min(i+76, len(enc))
		w.PrintfLine("%s", enc[i:end])
	}
}

// randomString returns a hex string of length 2*n (because each byte => two hex chars).
func randomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}

// generateMsgID generates a message id using the provided reference (domain or email).
// If ref has an @, use the right-hand side as domain, otherwise use ref as domain.
func generateMsgID(ref string) string {
	s := strings.Split(ref, "@")
	return fmt.Sprintf("%s@%s", randomString(16), s[len(s)-1])
}
