package smtp

import (
	"bytes"
	"errors"
	"fmt"
	"net/smtp"
	"slices"
)

// loginAuth implements the LOGIN authentication mechanism for SMTP.
var _ smtp.Auth = &loginAuth{}

// loginAuth holds the credentials and server information for LOGIN authentication.
type loginAuth struct {
	username, password, host string
}

// Start initiates the LOGIN authentication process, verifying TLS and server name.
func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS && !isLocalhost(server.Name) {
		return "", nil, errors.New("unencrypted connection")
	}
	if server.Name != a.host {
		return "", nil, errors.New("wrong server name")
	}
	resp := []byte(a.username)
	return "LOGIN", resp, nil
}

// Next handles the server's challenge, responding with username or password as needed.
func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		if bytes.Contains(fromServer, []byte("Username")) {
			return []byte(a.username), nil
		}
		if bytes.Contains(fromServer, []byte("Password")) {
			return []byte(a.password), nil
		}
		// We've already sent everything.
		return nil, fmt.Errorf("unexpected server challenge: %s", string(fromServer))
	}
	return nil, nil
}

// isLocalhost checks if the server name is a localhost address.
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
	if slices.Contains(c.auth, "CRAM-MD5") {
		a = smtp.CRAMMD5Auth(auth.Username, auth.Password)
	} else if slices.Contains(c.auth, "PLAIN") {
		a = smtp.PlainAuth(auth.Identity, auth.Username, auth.Password, auth.Server)
	} else {
		a = &loginAuth{auth.Username, auth.Password, auth.Server}

	}
	return c.Auth(a)
}
