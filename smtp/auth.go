package smtp

import (
	"errors"
	"net/smtp"
	"strings"
)

var _ smtp.Auth = &loginAuth{}

type loginAuth struct {
	username, password, server string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS && !isLocalhost(server.Name) {
		return "", nil, errors.New("unencrypted connection")
	}
	if server.Name != a.server {
		return "", nil, errors.New("wrong server name")
	}
	resp := []byte(a.username)
	return "LOGIN", resp, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		if strings.Contains(string(fromServer), "Username") {
			resp := []byte(a.username)
			return resp, nil
		}
		if strings.Contains(string(fromServer), "Password") {
			resp := []byte(a.password)
			return resp, nil
		}
		// We've already sent everything.
		return nil, errors.New("unexpected server challenge")
	}
	return nil, nil
}

func isLocalhost(name string) bool {
	return name == "localhost" || name == "127.0.0.1" || name == "::1"
}

// Auth is an SMTP authentication information.
type Auth struct {
	Identity string
	Username string
	Password string
	Server   string
}

// Auth2 authenticates a client using the provided authentication information.
// A failed authentication closes the connection.
// Only servers that advertise the AUTH extension support this function.
func (c *Client) Auth2(auth *Auth) error {
	// Auto select auth mode
	var a smtp.Auth
	if auths := strings.Join(c.auth, " "); strings.Contains(auths, "CRAM-MD5") {
		a = smtp.CRAMMD5Auth(auth.Username, auth.Password)
	} else if strings.Contains(auths, "PLAIN") {
		a = smtp.PlainAuth(auth.Identity, auth.Username, auth.Password, auth.Server)
	} else {
		a = &loginAuth{auth.Username, auth.Password, auth.Server}

	}

	return c.Auth(a)
}
