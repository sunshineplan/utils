package utils

import (
	"reflect"
)

// Deduplicate removes duplicate items in slice
func Deduplicate(slice interface{}) interface{} {
	if reflect.TypeOf(slice).Kind() != reflect.Slice {
		panic("only support slice arg")
	}
	items := reflect.ValueOf(slice)
	if items.Len() == 0 {
		return slice
	}
	t := reflect.TypeOf(items.Index(0).Interface())
	unique := reflect.MakeSlice(reflect.SliceOf(t), 0, 0)
	keys := make(map[interface{}]bool)
	for i := 0; i < items.Len(); i++ {
		if _, ok := keys[items.Index(i).Interface()]; !ok {
			keys[items.Index(i).Interface()] = true
			unique = reflect.Append(unique, items.Index(i))
		}
	}
	return unique.Interface()
}
