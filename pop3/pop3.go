package pop3

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/textproto"
	"strconv"
	"strings"
)

type Client struct {
	*textproto.Conn
}

const (
	lineBreak = "\r\n"

	respOK       = "+OK"
	respOKInfo   = "+OK "
	respErr      = "-ERR"
	respErrInfo  = "-ERR "
	respContinue = "+ "
)

func Dial(ctx context.Context, addr string) (*Client, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewClient(conn)
}

func DialTLS(ctx context.Context, addr string) (*Client, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	d := &tls.Dialer{Config: &tls.Config{ServerName: host}}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewClient(conn)
}

func NewClient(conn net.Conn) (*Client, error) {
	c := &Client{textproto.NewConn(conn)}
	s, err := c.ReadLine()
	if err != nil {
		return nil, err
	}
	slog.Debug("<<< " + s)

	if _, err = parseResp(s); err != nil {
		c.Close()
		return nil, err
	}

	return c, nil
}

// Stat returns the number of messages and their total size in bytes in the inbox.
func (c *Client) Stat() (count int, size int, err error) {
	s, err := c.Cmd("STAT", false)
	if err != nil {
		return
	}

	// count size
	f := strings.Fields(s)

	// Total number of messages.
	count, err = strconv.Atoi(f[0])
	if err != nil {
		return
	}
	if count == 0 {
		return
	}

	// Total size of all messages in bytes.
	size, err = strconv.Atoi(f[1])

	return
}

// MessageID contains the ID and size of an individual message.
type MessageID struct {
	// ID is the numerical index (non-unique) of the message.
	ID   int
	Size int

	// UID is only present if the response is to the UIDL command.
	UID string
}

// List returns a list of (message ID, message Size) pairs.
// If the optional id > 0, then only that particular message is listed.
// The message IDs are sequential, 1 to N.
func (c *Client) List(id int) ([]MessageID, error) {
	var (
		s   string
		err error
	)

	if id > 0 {
		// Single line response listing one message.
		s, err = c.Cmd("LIST %d", false, id)
	} else {
		// Multiline response listing all messages.
		s, err = c.Cmd("LIST", true)
	}
	if err != nil {
		return nil, err
	}

	var (
		out   []MessageID
		lines = strings.Split(s, lineBreak)
	)

	for _, l := range lines {
		// id size
		f := strings.Fields(l)
		if len(f) == 0 {
			break
		}

		id, err := strconv.Atoi(f[0])
		if err != nil {
			return nil, err
		}

		size, err := strconv.Atoi(f[1])
		if err != nil {
			return nil, err
		}

		out = append(out, MessageID{ID: id, Size: size})
	}

	return out, nil
}

// Uidl returns a list of (message ID, message UID) pairs. If the optional msgID
// is > 0, then only that particular message is listed. It works like Top() but only works on
// servers that support the UIDL command. Messages size field is not available in the UIDL response.
func (c *Client) Uidl(id int) ([]MessageID, error) {
	var (
		s   string
		err error
	)

	if id > 0 {
		// Single line response listing one message.
		s, err = c.Cmd("UIDL %d", false, id)
	} else {
		// Multiline response listing all messages.
		s, err = c.Cmd("UIDL", true)
	}
	if err != nil {
		return nil, err
	}

	var (
		out   []MessageID
		lines = strings.Split(s, lineBreak)
	)

	for _, l := range lines {
		// id uid
		f := strings.Fields(l)
		if len(f) == 0 {
			break
		}

		id, err := strconv.Atoi(f[0])
		if err != nil {
			return nil, err
		}

		out = append(out, MessageID{ID: id, UID: f[1]})
	}

	return out, nil
}

// Retr downloads a message by the given id and returns the data
// of the entire message.
func (c *Client) Retr(id int) (string, error) {
	return c.Cmd("RETR %d", true, id)
}

// Top retrieves a message by its ID with full headers and numLines lines of the body.
func (c *Client) Top(id int, numLines int) (string, error) {
	return c.Cmd("TOP %d %d", true, id, numLines)
}

// Dele deletes one or more messages. The server only executes the
// deletions after a successful Quit().
func (c *Client) Dele(ids ...int) error {
	for _, id := range ids {
		if _, err := c.Cmd("DELE %d", false, id); err != nil {
			return err
		}
	}
	return nil
}

// Rset clears the messages marked for deletion in the current session.
func (c *Client) Rset() error {
	_, err := c.Cmd("RSET", false)
	return err
}

// Noop issues a do-nothing NOOP command to the server. This is useful for
// prolonging open connections.
func (c *Client) Noop() error {
	_, err := c.Cmd("NOOP", false)
	return err
}

// Quit sends the QUIT command to server and gracefully closes the connection.
// Message deletions (DELE command) are only excuted by the server on a graceful
// quit and close.
func (c *Client) Quit() error {
	if _, err := c.Cmd("QUIT", false); err != nil {
		return err
	}
	return c.Close()
}

func (c *Client) Cmd(s string, isMulti bool, args ...any) (string, error) {
	slog.Debug(">>> " + fmt.Sprintf(s, args...))
	if _, err := c.Conn.Cmd(s, args...); err != nil {
		return "", err
	}

	s, err := c.ReadLine()
	if err != nil {
		return "", err
	}
	slog.Debug("<<< " + s)

	s, err = parseResp(s)
	if err != nil {
		return "", err
	}
	if !isMulti {
		return s, nil
	}

	var res string
	for {
		s, err := c.ReadLine()
		if err != nil {
			return "", err
		}
		slog.Debug("<<< " + s)
		if s == "." {
			break
		}

		res += s + lineBreak
	}

	return res, nil
}

func parseResp(s string) (string, error) {
	if len(s) == 0 {
		return "", nil
	}

	if s == respOK {
		return "", nil
	} else if strings.HasPrefix(s, respOKInfo) {
		return strings.TrimPrefix(s, respOKInfo), nil
	} else if s == respErr {
		return "", errors.New("unknown error (no info specified in response)")
	} else if strings.HasPrefix(s, respErrInfo) {
		return "", errors.New(strings.TrimPrefix(s, respErrInfo))
	} else if strings.HasPrefix(s, respContinue) {
		return strings.TrimPrefix(s, respContinue), nil
	} else {
		return "", fmt.Errorf("unknown response: %q", s)
	}
}
