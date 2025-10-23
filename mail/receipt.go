package mail

import (
	"encoding"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
)

var (
	_ encoding.TextUnmarshaler = new(Receipts)
	_ encoding.TextMarshaler   = Receipts{}
	_ json.Unmarshaler         = new(Receipts)
	_ json.Marshaler           = Receipts{}
)

// Receipt creates a mail.Address pointer from name and address.
func Receipt(name, address string) *mail.Address {
	return &mail.Address{Name: name, Address: address}
}

// Receipts represents a list of mail addresses.
type Receipts []*mail.Address

// ParseReceipts parses a comma/semicolon separated list of addresses into Receipts.
func ParseReceipts(rcpts string) (Receipts, error) {
	addresses, err := mail.ParseAddressList(rcpts)
	if err != nil {
		return nil, err
	}
	return Receipts(addresses), nil
}

// List returns a slice of address strings (just the email addresses).
func (rcpts Receipts) List() []string {
	s := make([]string, len(rcpts))
	for i, rcpt := range rcpts {
		s[i] = rcpt.Address
	}
	return s
}

// String returns the addresses joined in the standard RFC 5322 way
// using the String method of mail.Address.
func (rcpts Receipts) String() string {
	if len(rcpts) == 0 {
		return ""
	}
	var b strings.Builder
	// approximate average length to avoid repeated allocations
	b.Grow(len(rcpts) * 32)
	for i, rcpt := range rcpts {
		if i != 0 {
			b.WriteString(", ")
		}
		fmt.Fprint(&b, rcpt)
	}
	return b.String()
}

// UnmarshalText implements encoding.TextUnmarshaler
func (rcpts *Receipts) UnmarshalText(text []byte) error {
	addresses, err := ParseReceipts(string(text))
	if err != nil {
		return err
	}
	*rcpts = addresses
	return nil
}

// UnmarshalJSON supports either a single string (comma separated) or an array of strings.
func (rcpts *Receipts) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		return rcpts.UnmarshalText([]byte(s))
	}
	var list []string
	if err := json.Unmarshal(b, &list); err != nil {
		return err
	}
	var addresses []*mail.Address
	for _, i := range list {
		address, err := mail.ParseAddress(i)
		if err != nil {
			return err
		}
		addresses = append(addresses, address)
	}
	*rcpts = addresses
	return nil
}

// MarshalText implements encoding.TextMarshaler and returns the RFC5322 representation.
func (rcpts Receipts) MarshalText() ([]byte, error) {
	return []byte(rcpts.String()), nil
}

// MarshalJSON implements json.Marshaler.
// It encodes the receipts as a JSON array of email address strings.
func (rcpts Receipts) MarshalJSON() ([]byte, error) {
	return json.Marshal(rcpts.String())
}
