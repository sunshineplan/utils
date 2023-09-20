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

func (d *Dialer) Dial() (*smtp.Client, error) {
	if d.Timeout == 0 {
		d.Timeout = 3 * time.Minute
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout)
	defer cancel()

	client, err := smtp.Dial(ctx, fmt.Sprintf("%s:%d", d.Server, d.Port))
	if err != nil {
		return nil, err
	}
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(nil); err != nil {
			client.Quit()
			return nil, err
		}
	}
	if ok, _ := client.Extension("AUTH"); ok && d.Account != "" && d.Password != "" {
		if err = client.Auth2(&smtp.Auth{Identity: "", Username: d.Account, Password: d.Password, Server: d.Server}); err != nil {
			client.Quit()
			return nil, err
		}
	}

	return client, nil
}

// Send sends the given messages.
func (d *Dialer) Send(msg ...*Message) error {
	client, err := d.Dial()
	if err != nil {
		return err
	}
	defer client.Quit()

	for _, m := range msg {
		if m.From == nil {
			m.From = Receipt("", d.Account)
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
		go func() { c <- client.SendMail(m.From.Address, m.RcptList(), m.Bytes(d.Account)) }()

		select {
		case <-time.After(d.Timeout):
			return context.DeadlineExceeded
		case err := <-c:
			if err != nil {
				return err
			}
		}
	}

	return nil
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
