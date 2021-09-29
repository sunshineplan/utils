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

type httpProxy struct {
	*url.URL
	forward proxy.Dialer
}

func NewDialer(u *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	p := new(httpProxy)
	p.URL = u
	if forward == nil {
		forward = proxy.Direct
	}
	p.forward = forward

	return p, nil
}

func (p *httpProxy) dialForward() (net.Conn, error) {
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

func (p *httpProxy) Dial(network, addr string) (net.Conn, error) {
	conn, err := p.dialForward()
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: make(http.Header),
	}
	if p.User != nil {
		password, _ := p.User.Password()
		req.Header.Set("Proxy-Authorization", "Basic "+basicAuth(p.User.Username(), password))
	}

	resp, err := doRequest(conn, req)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if resp.StatusCode == http.StatusProxyAuthRequired && p.User != nil {
		password, _ := p.User.Password()
		req.SetBasicAuth(p.User.Username(), password)
		resp, err = doRequest(conn, req)
		if err != nil {
			conn.Close()
			return nil, err
		}
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

func doRequest(conn net.Conn, req *http.Request) (*http.Response, error) {
	if err := req.Write(conn); err != nil {
		return nil, err
	}

	br := bufio.NewReader(conn)
	return http.ReadResponse(br, req)
}

func init() {
	proxy.RegisterDialerType("http", NewDialer)
	proxy.RegisterDialerType("https", NewDialer)
}
