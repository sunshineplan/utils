package httpproxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
)

// A Dialer is a means to establish a connection.
type Dialer interface {
	// Dial connects to the given address via the proxy.
	Dial(network, addr string) (c net.Conn, err error)
}

// A ContextDialer dials using a context.
type ContextDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type Proxy struct {
	*url.URL
	forward Dialer
}

func New(u *url.URL, forward Dialer) *Proxy {
	p := &Proxy{URL: u}
	if forward == nil {
		forward = Direct
	}
	p.forward = forward

	return p
}

func NewDialer(u *url.URL, forward Dialer) (Dialer, error) {
	return New(u, forward), nil
}

func NewProxyDialer(u *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	return NewDialer(u, forward)
}

func (p *Proxy) dialForward() (net.Conn, error) {
	addr := p.Host

	conn, err := p.forward.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	if p.Scheme == "https" {
		colonPos := strings.LastIndex(addr, ":")
		if colonPos == -1 {
			colonPos = len(addr)
		}
		hostname := addr[:colonPos]

		conn = tls.Client(conn, &tls.Config{ServerName: hostname})
	}

	return conn, nil
}

func (p *Proxy) DialWithHeader(addr string, header http.Header) (net.Conn, *http.Response, error) {
	conn, err := p.dialForward()
	if err != nil {
		return nil, nil, err
	}

	req := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: header,
	}
	if err := req.Write(conn); err != nil {
		conn.Close()
		return nil, nil, err
	}

	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return conn, resp, nil
}

func (p *Proxy) Dial(network, addr string) (net.Conn, error) {
	if network != "tcp" {
		return nil, errors.New("network must be tcp")
	}

	header := make(http.Header)
	if p.User != nil {
		password, _ := p.User.Password()
		header.Set("Proxy-Authorization", "Basic "+basicAuth(p.User.Username(), password))
	}

	conn, resp, err := p.DialWithHeader(addr, header)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		conn.Close()
		return nil, errors.New(resp.Status)
	}

	return conn, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func init() {
	proxy.RegisterDialerType("http", NewProxyDialer)
	proxy.RegisterDialerType("https", NewProxyDialer)
}
