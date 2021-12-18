package smtp

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/smtp"
	"strings"
)

// A Client represents a client connection to an SMTP server.
type Client struct {
	*smtp.Client
	serverName string
}

// Dial returns a new Client connected to an SMTP server at addr.
// The addr must include a port, as in "mail.example.com:smtp".
func Dial(ctx context.Context, addr string) (*Client, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	return NewClient(conn, host)
}

// NewClient returns a new Client using an existing connection and host as a
// server name to be used when authenticating.
func NewClient(conn net.Conn, host string) (*Client, error) {
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return nil, err
	}
	return &Client{c, host}, nil
}

// Cmd is a convenience function that sends a command and returns the response.
func (c *Client) Cmd(expectCode int, format string, args ...interface{}) (int, string, error) {
	log.Printf("CMD: "+format, args...)
	id, err := c.Text.Cmd(format, args...)
	if err != nil {
		return 0, "", err
	}
	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)
	code, msg, err := c.Text.ReadResponse(expectCode)
	log.Println(code, msg)
	return code, msg, err
}

// Auth is an SMTP authentication information.
type Auth struct {
	Identity string
	Username string
	Password string
	Server   string
}

// Auth authenticates a client using the provided authentication mechanism.
// A failed authentication closes the connection.
// Only servers that advertise the AUTH extension support this function.
func (c *Client) Auth(auth *Auth) error {
	if auth == nil {
		return errors.New("auth is nil")
	}

	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: c.serverName}
		if err := c.StartTLS(config); err != nil {
			return err
		}
	}

	// Auto select auth mode
	ok, auths := c.Extension("AUTH")
	if !ok {
		return errors.New("smtp: server doesn't support AUTH")
	}

	var a smtp.Auth
	if strings.Contains(auths, "CRAM-MD5") {
		a = smtp.CRAMMD5Auth(auth.Username, auth.Password)
	} else if strings.Contains(auths, "PLAIN") {
		a = smtp.PlainAuth(auth.Identity, auth.Username, auth.Password, auth.Server)
	} else {
		a = &loginAuth{auth.Username, auth.Password, auth.Server}

	}

	return c.Client.Auth(a)
}

// Send sends an email from address from, to addresses to, with
// message msg.
func (c *Client) Send(from string, to []string, msg []byte) error {
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err := c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		return err
	}

	return w.Close()
}

// SendMail sends an email from address from, to addresses to, with
// message msg and quit.
func (c *Client) SendMail(from string, to []string, msg []byte) error {
	if err := c.Send(from, to, msg); err != nil {
		return err
	}
	return c.Quit()
}

// SendMail connects to the server at addr, switches to TLS if
// possible, authenticates with the optional mechanism a if possible,
// and then sends an email from address from, to addresses to, with
// message msg.
// The addr must include a port, as in "mail.example.com:smtp".
//
// The addresses in the to parameter are the SMTP RCPT addresses.
//
// The msg parameter should be an RFC 822-style email with headers
// first, a blank line, and then the message body. The lines of msg
// should be CRLF terminated. The msg headers should usually include
// fields such as "From", "To", "Subject", and "Cc".  Sending "Bcc"
// messages is accomplished by including an email address in the to
// parameter but not including it in the msg headers.
func SendMail(ctx context.Context, addr string, auth *Auth, from string, to []string, msg []byte) error {
	if err := validateLine(from); err != nil {
		return err
	}
	for _, recp := range to {
		if err := validateLine(recp); err != nil {
			return err
		}
	}
	c, err := Dial(ctx, addr)
	if err != nil {
		return err
	}
	defer c.Close()

	if err = c.Auth(auth); err != nil {
		return err
	}

	return c.SendMail(from, to, msg)
}

// validateLine checks to see if a line has CR or LF as per RFC 5321
func validateLine(line string) error {
	if strings.ContainsAny(line, "\n\r") {
		return errors.New("smtp: A line must not contain CR or LF")
	}
	return nil
}
