package unit

import (
	"encoding"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	_ encoding.TextMarshaler   = ByteSize(0)
	_ encoding.TextUnmarshaler = new(ByteSize)
)

var byteSizeRegexp = regexp.MustCompile(`^(\d+(?:\.\d+)?) ?((?i)[KMGTPE]?B?)$`)

var (
	byteSizeStr = map[ByteSize]string{
		B:  "B",
		KB: "KB",
		MB: "MB",
		GB: "GB",
		TB: "TB",
		PB: "PB",
		EB: "EB",
	}
	strByteSize = map[string]ByteSize{
		"B": B,
		"K": KB,
		"M": MB,
		"G": GB,
		"T": TB,
		"P": PB,
		"E": EB,
	}
)

type ByteSize int64

const (
	B  ByteSize = 1
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
)

func ParseByteSize(s string) (ByteSize, error) {
	s = strings.TrimSpace(s)
	res := byteSizeRegexp.FindStringSubmatch(s)
	if len(res) != 3 {
		return 0, errors.New("invalid byte size syntax")
	}
	unit := strings.ToUpper(res[2])
	if unit = strings.TrimSuffix(unit, "B"); unit == "" {
		unit = "B"
	}
	v, err := strconv.ParseFloat(res[1], 64)
	if err != nil {
		return 0, err
	}
	return ByteSize(v * float64(strByteSize[unit])), nil
}

func NewByteSize(n float64, unit ByteSize) ByteSize {
	return ByteSize(n * float64(unit))
}

func DefaultByteSizeFormatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}

var ByteSizeFormatFloat = DefaultByteSizeFormatFloat

func (n ByteSize) String() string {
	unit := B
	switch {
	case n >= EB:
		unit = EB
	case n >= PB:
		unit = PB
	case n >= TB:
		unit = TB
	case n >= GB:
		unit = GB
	case n >= MB:
		unit = MB
	case n >= KB:
		unit = KB
	}
	return ByteSizeFormatFloat(float64(n)/float64(unit)) + byteSizeStr[unit]
}

func (b ByteSize) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b *ByteSize) UnmarshalText(text []byte) error {
	bytes, err := ParseByteSize(string(text))
	if err != nil {
		return err
	}
	*b = bytes
	return nil
}
