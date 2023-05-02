package mail

import (
	"encoding"
	"fmt"
	"net/mail"
	"strings"
)

var (
	_ encoding.TextUnmarshaler = (*Receipts)(nil)
	_ encoding.TextMarshaler   = Receipts{}
)

func Receipt(name, address string) *mail.Address {
	return &mail.Address{Name: name, Address: address}
}

type Receipts []*mail.Address

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
	addresses, err := mail.ParseAddressList(string(text))
	if err != nil {
		return err
	}
	*rcpts = addresses
	return nil
}

func (rcpts Receipts) MarshalText() ([]byte, error) {
	return []byte(rcpts.String()), nil
}
