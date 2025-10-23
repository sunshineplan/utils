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

// ByteSize represents a quantity of bytes.
type ByteSize int64

// Common byte size units.
const (
	B  ByteSize = 1
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
)

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

// ParseByteSize parses a human-readable size string (e.g. "1.5GB", "100 KB")
// and returns the corresponding ByteSize value.
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

// MustParseByteSize parses a size string and panics if it is invalid.
func MustParseByteSize(s string) ByteSize {
	v, err := ParseByteSize(s)
	if err != nil {
		panic(err)
	}
	return v
}

// NewByteSize creates a ByteSize from a numeric value and a unit string.
// Example: NewByteSize(1.5, "GB") -> 1.5GB.
// Returns an error if the unit is not recognized.
func NewByteSize(n float64, unit string) (ByteSize, error) {
	unit = strings.ToUpper(strings.TrimSpace(unit))
	unit = strings.TrimSuffix(unit, "B")
	if unit == "" {
		unit = "B"
	}
	bs, ok := strByteSize[unit]
	if !ok {
		return 0, errors.New("unknown byte size unit: " + unit)
	}
	return ByteSize(n * float64(bs)), nil
}

// DefaultByteSizeFormatFloat formats a float with up to 2 decimals,
// trimming trailing zeros and the decimal point if unnecessary.
func DefaultByteSizeFormatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}

// ByteSizeFormatFloat defines how floating-point values are formatted in String().
// It can be replaced to customize output precision.
var ByteSizeFormatFloat = DefaultByteSizeFormatFloat

// String returns a human-readable representation of the byte size.
// e.g. 1536 -> "1.5KB", 1048576 -> "1MB".
func (n ByteSize) String() string {
	units := []ByteSize{EB, PB, TB, GB, MB, KB, B}
	for _, unit := range units {
		if n >= unit {
			return ByteSizeFormatFloat(float64(n)/float64(unit)) + byteSizeStr[unit]
		}
	}
	return "0B"
}

// MarshalText implements the encoding.TextMarshaler interface.
func (b ByteSize) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (b *ByteSize) UnmarshalText(text []byte) error {
	bytes, err := ParseByteSize(string(text))
	if err != nil {
		return err
	}
	*b = bytes
	return nil
}

// To converts the ByteSize to the specified unit and returns its string representation.
func (b ByteSize) To(unit string, decimals int) (string, error) {
	unit = strings.ToUpper(strings.TrimSpace(unit))
	unit = strings.TrimSuffix(unit, "B")
	if unit == "" {
		unit = "B"
	}
	base, ok := strByteSize[unit]
	if !ok {
		return "", errors.New("invalid unit: " + unit)
	}
	value := float64(b) / float64(base)
	s := strconv.FormatFloat(value, 'f', decimals, 64)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".") + byteSizeStr[base], nil
}
