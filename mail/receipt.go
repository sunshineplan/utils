package mail

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
)

var (
	_ json.Unmarshaler = (*Receipts)(nil)
	_ json.Marshaler   = Receipts{}
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

func (rcpts *Receipts) UnmarshalJSON(text []byte) error {
	if unquote, err := strconv.Unquote(string(text)); err == nil {
		addresses, err := mail.ParseAddressList(unquote)
		if err != nil {
			return err
		}
		*rcpts = addresses
		return nil
	}
	var s []string
	if err := json.Unmarshal(text, &s); err != nil {
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

func (rcpts Receipts) MarshalJSON() ([]byte, error) {
	return []byte(rcpts.String()), nil
}
