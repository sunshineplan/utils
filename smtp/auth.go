package smtp

import (
	"errors"
	"net/smtp"
	"strings"
)

var _ smtp.Auth = &loginAuth{}

type loginAuth struct {
	username, password, host string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS && !isLocalhost(server.Name) {
		return "", nil, errors.New("unencrypted connection")
	}
	if server.Name != a.host {
		return "", nil, errors.New("wrong host name")
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
