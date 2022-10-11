package csv

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

var errNilPtr = errors.New("destination pointer is nil") // embedded in descriptive error

func marshalText(i any) (text string, err error) {
	var b []byte
	switch v := i.(type) {
	case nil:
	case string:
		text = v
	case encoding.TextMarshaler:
		b, err = v.MarshalText()
		text = string(b)
	default:
		b, err = json.Marshal(v)
		text = string(b)
	}
	return
}

func convert(src string) any {
	if i64, err := strconv.ParseInt(src, 10, 64); err == nil {
		return i64
	}
	if f64, err := strconv.ParseFloat(src, 64); err == nil {
		return f64
	}
	if bv, err := strconv.ParseBool(src); err == nil {
		return bv
	}
	if regexp.MustCompile(`^\s*\[.*\]\s*$`).MatchString(src) {
		return json.RawMessage(src)
	}
	return src
}

// convertAssign copies to dest the value in src, converting it if possible.
// An error is returned if the copy would result in loss of information.
// dest should be a pointer type.
// https://golang.org/src/database/sql/convert.go?h=convertAssignRows#L219
func convertAssign(dest any, src string) error {
	// Common cases, without reflect.
	switch d := dest.(type) {
	case *string:
		if d == nil {
			return errNilPtr
		}
		*d = src
		return nil
	case *[]byte:
		if d == nil {
			return errNilPtr
		}
		*d = []byte(src)
		return nil
	case *bool:
		if d == nil {
			return errNilPtr
		}
		bv, err := strconv.ParseBool(src)
		if err == nil {
			*d = bv
		}
		return err
	case *any:
		if d == nil {
			return errNilPtr
		}
		*d = src
		return nil
	case json.Unmarshaler:
		if d == nil {
			return errNilPtr
		}
		return d.UnmarshalJSON([]byte(src))
	case encoding.TextUnmarshaler:
		if d == nil {
			return errNilPtr
		}
		return d.UnmarshalText([]byte(src))
	}

	dpv := reflect.ValueOf(dest)
	if dpv.Kind() != reflect.Ptr {
		return errors.New("destination not a pointer")
	}
	if dpv.IsNil() {
		return errNilPtr
	}

	sv := reflect.ValueOf(src)
	dv := reflect.Indirect(dpv)
	if sv.Type().AssignableTo(dv.Type()) {
		dv.Set(sv)
		return nil
	}

	// The following conversions use a string value as an intermediate representation
	// to convert between various numeric types.
	//
	// This also allows scanning into user defined types such as "type Int int64".
	// For symmetry, also check for string destination types.
	if src == "" {
		return fmt.Errorf("converting Empty String to %s is unsupported", dv.Kind())
	}
	switch dv.Kind() {
	case reflect.Ptr:
		dv.Set(reflect.New(dv.Type().Elem()))
		return convertAssign(dv.Interface(), src)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64, err := strconv.ParseInt(src, 10, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting type String (%q) to a %s: %v", src, dv.Kind(), err)
		}
		dv.SetInt(i64)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u64, err := strconv.ParseUint(src, 10, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting type String (%q) to a %s: %v", src, dv.Kind(), err)
		}
		dv.SetUint(u64)
		return nil
	case reflect.Float32, reflect.Float64:
		f64, err := strconv.ParseFloat(src, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting type String (%q) to a %s: %v", src, dv.Kind(), err)
		}
		dv.SetFloat(f64)
		return nil
	}

	return json.Unmarshal([]byte(src), dest)
}

func strconvErr(err error) error {
	if numError, ok := err.(*strconv.NumError); ok {
		return numError.Err
	}
	return err
}
