// https://github.com/scorredoira/email
package mail

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"mime"
	"net/mail"
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
	buf.WriteString("From: " + m.From + "\r\n")

	t := time.Now()
	buf.WriteString("Date: " + t.Format(time.RFC1123Z) + "\r\n")

	buf.WriteString("To: " + strings.Join(m.To, ",") + "\r\n")
	if len(m.Cc) > 0 {
		buf.WriteString("Cc: " + strings.Join(m.Cc, ",") + "\r\n")
	}

	var coder = base64.StdEncoding
	var subject = "=?UTF-8?B?" + coder.EncodeToString([]byte(m.Subject)) + "?="
	buf.WriteString("Subject: " + subject + "\r\n")

	buf.WriteString("MIME-Version: 1.0\r\n")

	boundary := randomBoundary()

	if len(m.Attachments) > 0 {
		buf.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n")
		buf.WriteString("\r\n--" + boundary + "\r\n")
	}

	buf.WriteString(fmt.Sprintf("Content-Type: %s; charset=utf-8\r\n\r\n", m.ContentType))
	buf.WriteString(m.Body)
	buf.WriteString("\r\n")

	if len(m.Attachments) > 0 {
		for _, attachment := range m.Attachments {
			buf.WriteString("\r\n\r\n--" + boundary + "\r\n")

			if attachment.Inline {
				buf.WriteString("Content-Type: message/rfc822\r\n")
				buf.WriteString("Content-Disposition: inline; filename=\"" + attachment.Filename + "\"\r\n\r\n")

				buf.Write(attachment.Bytes)
			} else {
				ext := filepath.Ext(attachment.Filename)
				mimetype := mime.TypeByExtension(ext)
				if mimetype != "" {
					buf.WriteString(fmt.Sprintf("Content-Type: %s\r\n", mimetype))
				} else {
					buf.WriteString("Content-Type: application/octet-stream\r\n")
				}
				buf.WriteString("Content-Transfer-Encoding: base64\r\n")

				buf.WriteString("Content-Disposition: attachment; filename=\"=?UTF-8?B?")
				buf.WriteString(coder.EncodeToString([]byte(attachment.Filename)))
				buf.WriteString("?=\"\r\n\r\n")

				b := make([]byte, base64.StdEncoding.EncodedLen(len(attachment.Bytes)))
				base64.StdEncoding.Encode(b, attachment.Bytes)

				// write base64 content in lines of up to 76 chars
				for i, l := 0, len(b); i < l; i++ {
					buf.WriteByte(b[i])
					if (i+1)%76 == 0 {
						buf.WriteString("\r\n")
					}
				}
			}

			buf.WriteString("\r\n--" + boundary)
		}

		buf.WriteString("--")
	}

	return buf.Bytes()
}

func randomBoundary() string {
	b := make([]byte, 30)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}
