package httpproxy

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
)

type Proxy struct {
	*url.URL
	forward proxy.Dialer
}

func NewDialer(u *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	p := new(Proxy)
	p.URL = u
	if forward == nil {
		forward = proxy.Direct
	}
	p.forward = forward

	return p, nil
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
		return nil, fmt.Errorf("network must be tcp")
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
		return nil, fmt.Errorf("no StatusOK: [%d]", resp.StatusCode)
	}

	return conn, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func init() {
	proxy.RegisterDialerType("http", NewDialer)
	proxy.RegisterDialerType("https", NewDialer)
}
