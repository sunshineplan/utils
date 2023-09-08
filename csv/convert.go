package csv

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

var (
	errNilPtr = errors.New("destination pointer is nil") // embedded in descriptive error
	errNotSet = errors.New("not set")
)

func marshalText(i any) (text string, err error) {
	switch v := i.(type) {
	case nil:
	case string:
		text = v
	case encoding.TextMarshaler:
		var b []byte
		if b, err = v.MarshalText(); err == nil {
			text = string(b)
		}
	default:
		var b []byte
		if b, err = json.Marshal(v); err == nil {
			text = string(b)
		}
	}
	return
}

func strconvErr(err error) error {
	if numError, ok := err.(*strconv.NumError); ok {
		return numError.Err
	}
	return err
}

func convert(t reflect.Type, s string) (reflect.Value, error) {
	if t.Kind() == reflect.Ptr {
		return convert(t.Elem(), s)
	}
	v := reflect.Indirect(reflect.New(t))
	if v.CanInt() {
		n, err := strconv.ParseInt(s, 10, t.Bits())
		if err != nil {
			return v, fmt.Errorf("converting type String %q to a %s: %v", s, t.Kind(), strconvErr(err))
		}
		v.SetInt(n)
		return v, nil
	} else if v.CanUint() {
		u, err := strconv.ParseUint(s, 10, t.Bits())
		if err != nil {
			return v, fmt.Errorf("converting type String %q to a %s: %v", s, t.Kind(), strconvErr(err))
		}
		v.SetUint(u)
		return v, nil
	} else if v.CanFloat() {
		f, err := strconv.ParseFloat(s, t.Bits())
		if err != nil {
			return v, fmt.Errorf("converting type String %q to a %s: %v", s, t.Kind(), strconvErr(err))
		}
		v.SetFloat(f)
		return v, nil
	} else if v.CanComplex() {
		c, err := strconv.ParseComplex(s, t.Bits())
		if err != nil {
			return v, fmt.Errorf("converting type String %q to a %s: %v", s, t.Kind(), strconvErr(err))
		}
		v.SetComplex(c)
		return v, nil
	}
	switch t.Kind() {
	case reflect.String, reflect.Interface:
		v.SetString(s)
		return v, nil
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err == nil {
			return v, fmt.Errorf("converting type String %q to a %s: %v", s, t.Kind(), err)
		}
		v.SetBool(b)
		return v, nil
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			v.SetBytes([]byte(s))
			return v, nil
		}
	}
	return v, errNotSet
}

// setCell copies to dest the value in src, converting it if possible.
// An error is returned if the copy would result in loss of information.
// dest should be a pointer type.
// https://golang.org/src/database/sql/convert.go?h=convertAssignRows#L219
func setCell(dest any, s string) error {
	// Common cases, without reflect.
	switch d := dest.(type) {
	case *string:
		if d == nil {
			return errNilPtr
		}
		*d = s
		return nil
	case *[]byte:
		if d == nil {
			return errNilPtr
		}
		*d = []byte(s)
		return nil
	case *bool:
		if d == nil {
			return errNilPtr
		}
		bv, err := strconv.ParseBool(s)
		if err == nil {
			*d = bv
		}
		return err
	case *any:
		if d == nil {
			return errNilPtr
		}
		*d = s
		return nil
	case json.Unmarshaler:
		if d == nil {
			return errNilPtr
		}
		return d.UnmarshalJSON([]byte(s))
	case encoding.TextUnmarshaler:
		if d == nil {
			return errNilPtr
		}
		return d.UnmarshalText([]byte(s))
	}

	dpv := reflect.ValueOf(dest)
	if dpv.Kind() != reflect.Ptr {
		return errors.New("destination not a pointer")
	}
	if dpv.IsNil() {
		return errNilPtr
	}
	dv := reflect.Indirect(dpv)
	if v := reflect.ValueOf(s); v.Type().AssignableTo(dv.Type()) {
		dv.Set(v)
		return nil
	}
	if dv.Kind() == reflect.Pointer {
		dv.Set(reflect.New(dv.Type().Elem()))
		return setCell(dv.Interface(), s)
	}

	if v, err := convert(dv.Type(), s); err != nil {
		if err == errNotSet {
			return json.Unmarshal([]byte(s), dest)
		}
		return err
	} else {
		dv.Set(v)
		return nil
	}
}

func setRow(dest any, m map[string]string) error {
	// Common cases, without reflect.
	switch d := dest.(type) {
	case *map[string]string:
		if d == nil {
			return errNilPtr
		}
		*d = m
		return nil
	case json.Unmarshaler:
		if d == nil {
			return errNilPtr
		}
		b, _ := json.Marshal(m)
		return d.UnmarshalJSON(b)
	}

	dpv := reflect.ValueOf(dest)
	if dpv.Kind() != reflect.Ptr {
		return errors.New("destination not a pointer")
	}
	if dpv.IsNil() {
		return errNilPtr
	}

	dv := reflect.Indirect(dpv)
	if v := reflect.ValueOf(m); v.Type().AssignableTo(dv.Type()) {
		dv.Set(v)
		return nil
	}

	if len(m) == 0 {
		return nil
	}
	switch dv.Kind() {
	case reflect.Ptr:
		dv.Set(reflect.New(dv.Type().Elem()))
		return setRow(dv.Interface(), m)
	case reflect.Map:
		if dv.Type().Key().Kind() != reflect.String {
			return errors.New("map key is not string kind")
		}
		for k, v := range m {
			i := reflect.New(dv.Type().Elem())
			if err := setCell(i.Interface(), v); err != nil {
				return err
			}
			dv.SetMapIndex(reflect.ValueOf(k), reflect.Indirect(i))
		}
		return nil
	case reflect.Struct:
		csvTags := make(map[string]string)
		for _, field := range reflect.VisibleFields(dv.Type()) {
			if tag, ok := field.Tag.Lookup("csv"); ok {
				csvTags[tag] = field.Name
			}
		}
		for k, v := range m {
			var val reflect.Value
			if field, ok := csvTags[k]; ok {
				val = dv.FieldByName(field)
			} else {
				val = dv.FieldByName(k)
			}
			if val.IsValid() {
				i := reflect.New(val.Type())
				if err := setCell(i.Interface(), v); err != nil {
					return err
				}
				val.Set(reflect.Indirect(i))
			}
		}
		return nil
	}
	b, _ := json.Marshal(m)
	return json.Unmarshal(b, dest)
}
