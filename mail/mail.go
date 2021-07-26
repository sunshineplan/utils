package mail

import (
	"context"
	"os"
	"path/filepath"
	"time"
)

// Dialer is a dialer to an SMTP server.
type Dialer struct {
	Host     string
	Port     int
	Account  string
	Password string
	Timeout  time.Duration
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

// ContentType represents content type
type ContentType int

const (
	// TextPlain sets body type to text/plain in message body
	TextPlain ContentType = iota
	// TextHTML sets body type to text/html in message body
	TextHTML
)

// Attachment represents an email attachment
type Attachment struct {
	Filename string
	Path     string
	Bytes    []byte
	Inline   bool
}

// Send sends the given messages.
func (d *Dialer) Send(msg ...*Message) error {
	for _, m := range msg {
		if m.From == "" {
			m.From = d.Account
		}

		for _, i := range m.Attachments {
			if i.Bytes != nil {
				if i.Filename == "" {
					i.Filename = "attachment"
				}
			} else {
				data, err := os.ReadFile(i.Path)
				if err != nil {
					return err
				}

				i.Bytes = data
				if i.Filename == "" {
					i.Filename = filepath.Base(i.Path)
				}
			}
		}

		if d.Timeout == 0 {
			d.Timeout = 3 * time.Minute
		}

		ctx, cancel := context.WithTimeout(context.Background(), d.Timeout)
		defer cancel()

		c := make(chan error, 1)
		go func() { c <- d.sendMail(m.From, m.toList(), m.bytes()) }()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-c:
			if err != nil {
				return err
			}
		}
	}

	return nil
}
