package mail

import (
	"encoding"
	"encoding/json"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
)

var (
	_ encoding.TextUnmarshaler = (*Receipts)(nil)
	_ encoding.TextMarshaler   = Receipts{}
	_ json.Unmarshaler         = (*Receipts)(nil)
	_ json.Marshaler           = Receipts{}
)

func Receipt(name, address string) *mail.Address {
	return &mail.Address{Name: name, Address: address}
}

type Receipts []*mail.Address

func ParseReceipts(rcpts string) (Receipts, error) {
	addresses, err := mail.ParseAddressList(rcpts)
	if err != nil {
		return nil, err
	}
	return Receipts(addresses), nil
}

func (rcpts Receipts) List() []string {
	var s []string
	for _, rcpt := range rcpts {
		s = append(s, rcpt.Address)
	}
	return s
}

func (rcpts Receipts) String() string {
	var b strings.Builder
	for i, rcpt := range rcpts {
		if i != 0 {
			b.WriteString(", ")
		}
		fmt.Fprint(&b, rcpt)
	}
	return b.String()
}

func (rcpts *Receipts) UnmarshalText(text []byte) error {
	addresses, err := ParseReceipts(string(text))
	if err != nil {
		return err
	}
	*rcpts = addresses
	return nil
}

func (rcpts *Receipts) UnmarshalJSON(b []byte) error {
	if unquote, err := strconv.Unquote(string(b)); err == nil {
		return rcpts.UnmarshalText([]byte(unquote))
	}
	var s []string
	if err := json.Unmarshal(b, &s); err != nil {
		return nil
	}
	var addresses []*mail.Address
	for _, i := range s {
		address, err := mail.ParseAddress(i)
		if err != nil {
			return err
		}
		addresses = append(addresses, address)
	}
	*rcpts = addresses
	return nil
}

func (rcpts Receipts) MarshalText() ([]byte, error) {
	return []byte(rcpts.String()), nil
}

func (rcpts Receipts) MarshalJSON() ([]byte, error) {
	return []byte(rcpts.String()), nil
}
