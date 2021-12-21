package mail

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sunshineplan/utils/smtp"
)

// Dialer is a dialer to an SMTP server.
type Dialer struct {
	Server   string
	Port     int
	Account  string
	Password string
	Timeout  time.Duration
}

// Attachment represents an email attachment
type Attachment struct {
	Filename string
	Path     string
	Bytes    []byte
	Inline   bool
}

// SendMail connects to the server at Dialer's addr, switches to TLS if
// possible, authenticates with the optional mechanism a if possible,
// and then sends an email from address from, to addresses to, with
// message msg.
//
// The addresses in the to parameter are the SMTP RCPT addresses.
//
// The msg parameter should be an RFC 822-style email with headers
// first, a blank line, and then the message body. The lines of msg
// should be CRLF terminated. The msg headers should usually include
// fields such as "From", "To", "Subject", and "Cc".  Sending "Bcc"
// messages is accomplished by including an email address in the to
// parameter but not including it in the msg headers.
func (d *Dialer) SendMail(ctx context.Context, from string, to []string, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", d.Server, d.Port)
	auth := &smtp.Auth{Identity: "", Username: d.Account, Password: d.Password, Server: d.Server}
	return smtp.SendMail(ctx, addr, auth, from, to, msg)
}

// Send sends the given messages.
func (d *Dialer) Send(msg ...*Message) error {
	if d.Timeout == 0 {
		d.Timeout = 3 * time.Minute
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout)
	defer cancel()

	smtpClient, err := smtp.Dial(ctx, fmt.Sprintf("%s:%d", d.Server, d.Port))
	if err != nil {
		return err
	}
	defer smtpClient.Quit()

	if err = smtpClient.Auth(&smtp.Auth{Identity: "", Username: d.Account, Password: d.Password, Server: d.Server}); err != nil {
		return err
	}

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

		c := make(chan error, 1)
		go func() { c <- smtpClient.Send(m.From, m.RcptList(), m.Bytes()) }()

		select {
		case <-time.After(d.Timeout):
			return fmt.Errorf("timeout")
		case err := <-c:
			if err != nil {
				return err
			}
		}
	}

	return nil
}
