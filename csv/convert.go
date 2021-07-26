package csv

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

var errNilPtr = errors.New("destination pointer is nil") // embedded in descriptive error

// convertAssign copies to dest the value in src, converting it if possible.
// An error is returned if the copy would result in loss of information.
// dest should be a pointer type.
// https://golang.org/src/database/sql/convert.go?h=convertAssignRows#L219
func convertAssign(dest interface{}, src string) error {
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
		bv, err := strconv.ParseBool(src)
		if err == nil {
			*d = bv
		}
		return err
	case *interface{}:
		*d = src
		return nil
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
	case reflect.String:
		dv.SetString(src)
		return nil
	}

	return json.Unmarshal([]byte(src), dest)
}

func strconvErr(err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		return ne.Err
	}
	return err
}
