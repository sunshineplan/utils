// https://github.com/scorredoira/email
package mail

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding"
	"encoding/base64"
	"fmt"
	"math"
	"mime"
	"net/mail"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ParseAddress     = mail.ParseAddress
	ParseAddressList = mail.ParseAddressList
)

var (
	_ encoding.TextUnmarshaler = (*Receipts)(nil)
	_ encoding.TextMarshaler   = Receipts{}
)

type Receipts []*mail.Address

func (rcpts Receipts) String() string {
	var b strings.Builder
	for i, rcpt := range rcpts {
		if i != 0 {
			b.WriteString(", ")
		}
		fmt.Fprint(&b, rcpt)
	}
	return b.String()
}

func (rcpts *Receipts) UnmarshalText(text []byte) error {
	addresses, err := ParseAddressList(string(text))
	if err != nil {
		return err
	}
	*rcpts = addresses
	return nil
}

func (rcpts Receipts) MarshalText() ([]byte, error) {
	return []byte(rcpts.String()), nil
}

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
	From        string
	To, Cc, Bcc Receipts
	Subject     string
	Body        string
	ContentType ContentType
	Attachments []*Attachment
}

func (m *Message) RcptList() (rcpts []string) {
	for _, to := range m.To {
		rcpts = append(rcpts, to.Address)
	}
	for _, cc := range m.Cc {
		rcpts = append(rcpts, cc.Address)
	}
	for _, bcc := range m.Bcc {
		rcpts = append(rcpts, bcc.Address)
	}
	return
}

func (m *Message) Bytes(id string) []byte {
	if id == "" {
		e, err := ParseAddress(m.From)
		if err != nil {
			if hostname, _ := os.Hostname(); hostname != "" {
				id = hostname
			} else {
				id = "localhost"
			}
		} else {
			id = e.Address
		}
	}
	id = generateMsgID(id)

	var buf bytes.Buffer
	w := textproto.NewWriter(bufio.NewWriter(&buf))

	w.PrintfLine("MIME-Version: 1.0")
	w.PrintfLine("Date: " + time.Now().Format(time.RFC1123Z))
	w.PrintfLine("Message-ID: <%s>", id)
	w.PrintfLine("Subject: =?UTF-8?B?%s?=", toBase64(m.Subject))
	w.PrintfLine("From: " + m.From)
	w.PrintfLine("To: " + m.To.String())
	if len(m.Cc) > 0 {
		w.PrintfLine("Cc: " + m.Cc.String())
	}

	boundary := randomString(16)
	if len(m.Attachments) > 0 {
		w.PrintfLine("Content-Type: multipart/mixed; boundary=%q", boundary)
		w.PrintfLine("")
		w.PrintfLine("--" + boundary)
	}

	w.PrintfLine(`Content-Type: %s; charset="UTF-8"`, m.ContentType)
	w.PrintfLine("Content-Transfer-Encoding: base64")
	w.PrintfLine("")
	w.PrintfLine(toBase64(m.Body))

	if l := len(m.Attachments); l > 0 {
		for i, attachment := range m.Attachments {
			w.PrintfLine("--" + boundary)
			if attachment.ContentID != "" {
				w.PrintfLine("Content-Type: message/rfc822")
				w.PrintfLine(`Content-Disposition: inline; filename="=?UTF-8?B?%s?="`, toBase64(attachment.Filename))
				w.PrintfLine("Content-ID: <%s>", attachment.ContentID)
			} else {
				if mimetype := mime.TypeByExtension(filepath.Ext(attachment.Filename)); mimetype != "" {
					w.PrintfLine("Content-Type: " + mimetype)
				} else {
					w.PrintfLine("Content-Type: application/octet-stream")
				}
				w.PrintfLine(`Content-Disposition: attachment; filename="=?UTF-8?B?%s?="`, toBase64(attachment.Filename))
			}
			w.PrintfLine("Content-Transfer-Encoding: base64")
			w.PrintfLine("")

			b := make([]byte, base64.StdEncoding.EncodedLen(len(attachment.Bytes)))
			base64.StdEncoding.Encode(b, attachment.Bytes)

			// write base64 content in lines of up to 76 chars
			for i, l := 0, int(math.Ceil(float64(len(b))/76)); i < l; i++ {
				if i == l-1 {
					w.PrintfLine(string(b[i*76:]))
				} else {
					w.PrintfLine(string(b[i*76 : (i+1)*76]))
				}
			}

			if i < l-1 {
				w.PrintfLine("--" + boundary)
			} else {
				w.PrintfLine("--" + boundary + "--")
			}
		}
	}

	return buf.Bytes()
}

func toBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func randomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}

func generateMsgID(ref string) string {
	s := strings.Split(ref, "@")
	return fmt.Sprintf("%s@%s", randomString(16), s[len(s)-1])
}
