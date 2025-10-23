package pop3

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/textproto"
	"strconv"
	"strings"
)

// Client represents a POP3 client connection.
// It embeds a textproto.Conn for low-level protocol communication.
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

// Dial establishes a plain TCP connection to the POP3 server at the given address.
// The connection respects the provided context for timeout or cancellation.
func Dial(ctx context.Context, addr string) (*Client, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewClient(conn)
}

// DialTLS establishes a secure POP3-over-TLS connection to the given address.
// The server name is automatically derived from the address for certificate verification.
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

// NewClient initializes a POP3 client from an existing connection.
// It reads the server greeting line and validates that it starts with "+OK".
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

// Auth authenticates the user using the USER/PASS commands.
// Returns an error if either step fails.
func (c *Client) Auth(user, pass string) error {
	if _, err := c.Cmd("USER %s", false, user); err != nil {
		return err
	}
	_, err := c.Cmd("PASS %s", false, pass)
	return err
}

// Stat returns the number of messages and the total mailbox size (in bytes).
func (c *Client) Stat() (count int, size int, err error) {
	s, err := c.Cmd("STAT", false)
	if err != nil {
		return
	}
	f := strings.Fields(s)
	if len(f) < 2 {
		return 0, 0, fmt.Errorf("invalid STAT response: %q", s)
	}
	count, err = strconv.Atoi(f[0])
	if err != nil {
		return
	}
	if count == 0 {
		return
	}
	size, err = strconv.Atoi(f[1])
	return
}

// MessageID represents a single message entry as returned by LIST or UIDL.
// It includes the message index, size, and optional UID.
type MessageID struct {
	ID   int    // Numerical message index (1-based)
	Size int    // Message size in bytes
	UID  string // Optional UID (only for UIDL command)
}

// List returns message IDs and sizes from the mailbox.
// If id > 0, only that specific message is listed (single-line response).
// If id == 0, all messages are listed (multi-line response).
func (c *Client) List(id int) ([]MessageID, error) {
	var (
		s   string
		err error
	)

	if id > 0 {
		s, err = c.Cmd("LIST %d", false, id)
	} else {
		s, err = c.Cmd("LIST", true)
	}
	if err != nil {
		return nil, err
	}

	var out []MessageID
	for l := range strings.SplitSeq(s, lineBreak) {
		f := strings.Fields(l)
		if len(f) == 0 {
			continue
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

// Uidl returns message IDs and their unique identifiers (UIDs).
// If id > 0, only that specific message is listed.
// The UIDL command may not be supported by all servers.
func (c *Client) Uidl(id int) ([]MessageID, error) {
	var (
		s   string
		err error
	)

	if id > 0 {
		s, err = c.Cmd("UIDL %d", false, id)
	} else {
		s, err = c.Cmd("UIDL", true)
	}
	if err != nil {
		return nil, err
	}

	var out []MessageID
	for l := range strings.SplitSeq(s, lineBreak) {
		f := strings.Fields(l)
		if len(f) == 0 {
			continue
		}
		id, err := strconv.Atoi(f[0])
		if err != nil {
			return nil, err
		}
		out = append(out, MessageID{ID: id, UID: f[1]})
	}
	return out, nil
}

// Retr retrieves the full message text by ID, including headers and body.
func (c *Client) Retr(id int) (string, error) {
	return c.Cmd("RETR %d", true, id)
}

// Top retrieves message headers and the first numLines of the body.
func (c *Client) Top(id int, numLines int) (string, error) {
	return c.Cmd("TOP %d %d", true, id, numLines)
}

// Dele marks one or more messages for deletion.
// The deletions are finalized only after a successful Quit().
func (c *Client) Dele(ids ...int) error {
	for _, id := range ids {
		if _, err := c.Cmd("DELE %d", false, id); err != nil {
			return err
		}
	}
	return nil
}

// Rset resets the deletion marks on all messages in the current session.
func (c *Client) Rset() error {
	_, err := c.Cmd("RSET", false)
	return err
}

// Noop sends a NOOP command to keep the connection alive.
// Useful for preventing idle timeouts.
func (c *Client) Noop() error {
	_, err := c.Cmd("NOOP", false)
	return err
}

// Quit sends the QUIT command and closes the connection gracefully.
// Deletions are committed only if QUIT succeeds.
func (c *Client) Quit() error {
	if _, err := c.Cmd("QUIT", false); err != nil {
		c.Close()
		return err
	}
	return c.Close()
}

// Cmd sends a POP3 command with optional arguments and reads the response.
// If isMulti is true, the response is treated as multi-line and read until
// a line containing only "." is encountered. All lines are concatenated and returned.
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

	var b strings.Builder
	for {
		s, err := c.ReadLine()
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return "", err
		}
		slog.Debug("<<< " + s)

		// A single dot line marks the end of multi-line response.
		// Lines beginning with a dot have one dot removed as per POP3 spec.
		if len(s) > 0 && s[0] == '.' {
			if len(s) == 1 {
				break
			}
			s = s[1:]
		}
		b.WriteString(s)
		b.WriteString(lineBreak)
	}
	return b.String(), nil
}

// parseResp interprets a single-line POP3 response.
// It distinguishes between +OK, -ERR, and continuation ("+ ") responses,
// returning the response message or an error if the response is invalid.
func parseResp(s string) (string, error) {
	switch s {
	case "", respOK:
		return "", nil
	case respErr:
		return "", errors.New("server returned -ERR without info")
	default:
		switch {
		case strings.HasPrefix(s, respOKInfo):
			return strings.TrimPrefix(s, respOKInfo), nil
		case strings.HasPrefix(s, respErrInfo):
			return "", errors.New(strings.TrimPrefix(s, respErrInfo))
		case strings.HasPrefix(s, respContinue):
			// Some servers send "+ " for continuation prompts (rare in simple POP3).
			return strings.TrimPrefix(s, respContinue), nil
		default:
			return "", fmt.Errorf("unknown response: %q", s)
		}
	}
}
