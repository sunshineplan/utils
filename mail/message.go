// https://github.com/scorredoira/email
package mail

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	"mime"
	"net/mail"
	"net/textproto"
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

// Message represents an email message
type Message struct {
	From        string
	To, Cc, Bcc []string
	Subject     string
	Body        string
	ContentType ContentType
	Attachments []*Attachment
}

func (m *Message) RcptList() []string {
	rcptList := []string{}

	toList, _ := mail.ParseAddressList(strings.Join(m.To, ","))
	for _, to := range toList {
		rcptList = append(rcptList, to.Address)
	}

	ccList, _ := mail.ParseAddressList(strings.Join(m.Cc, ","))
	for _, cc := range ccList {
		rcptList = append(rcptList, cc.Address)
	}

	bccList, _ := mail.ParseAddressList(strings.Join(m.Bcc, ","))
	for _, bcc := range bccList {
		rcptList = append(rcptList, bcc.Address)
	}

	return rcptList
}

func (m *Message) Bytes() []byte {
	var buf bytes.Buffer
	w := textproto.NewWriter(bufio.NewWriter(&buf))

	w.PrintfLine("MIME-Version: 1.0")
	w.PrintfLine("Date: " + time.Now().Format(time.RFC1123Z))
	w.PrintfLine("Subject: =?UTF-8?B?%s?=", toBase64(m.Subject))
	w.PrintfLine("From: " + m.From)
	w.PrintfLine("To: " + strings.Join(m.To, ","))
	if len(m.Cc) > 0 {
		w.PrintfLine("Cc: " + strings.Join(m.Cc, ","))
	}

	boundary := randomBoundary()
	if len(m.Attachments) > 0 {
		w.PrintfLine("Content-Type: multipart/mixed; boundary=" + boundary)
		w.PrintfLine("")
		w.PrintfLine("--" + boundary)
	}

	w.PrintfLine("Content-Type: %s; charset=utf-8", m.ContentType)
	w.PrintfLine("Content-Transfer-Encoding: base64")
	w.PrintfLine("")
	w.PrintfLine(toBase64(m.Body))

	if l := len(m.Attachments); l > 0 {
		for i, attachment := range m.Attachments {
			w.PrintfLine("--" + boundary)
			if attachment.Inline {
				w.PrintfLine("Content-Type: message/rfc822")
				w.PrintfLine(`Content-Disposition: inline; filename="=?UTF-8?B?%s?="`, toBase64(attachment.Filename))
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

func randomBoundary() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}
