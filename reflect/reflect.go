package reflect

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

var (
	errValueIsNil           = errors.New("Invalid value, should be not nil")
	errValueIsNotPointer    = errors.New("Invalid value, should be pointer")
	errValueIsNotStruct     = errors.New("Invalid value, should be struct")
	errValueIsNotAssignable = errors.New("Invalid value, should be assignable")
)

// func getMarshaler(v reflect.Value) encoding.TextMarshaler {
// 	if v.CanInterface() {
// 		if m, ok := v.Interface().(encoding.TextMarshaler); ok {
// 			return m
// 		}
// 	}
// 	return nil
// }

// func findMarshaler(v reflect.Value) encoding.TextMarshaler {
// 	if m := getMarshaler(v); m != nil {
// 		return m
// 	}
// 	if v.CanAddr() {
// 		if m := getMarshaler(v.Addr()); m != nil {
// 			return m
// 		}
// 	}
// 	if m := getMarshaler(reflect.Indirect(v)); m != nil {
// 		return m
// 	}
// 	return nil
// }

// type TextMarshalers map[reflect.Type]func() (text []byte, err error)
// type TextUnmarshalers map[reflect.Type]func(text []byte) error

func getUnmarshaler(v reflect.Value) encoding.TextUnmarshaler {
	if v.CanInterface() {
		if u, ok := v.Interface().(encoding.TextUnmarshaler); ok {
			return u
		}
	}
	return nil
}

func findUnmarshaler(v reflect.Value) encoding.TextUnmarshaler {
	if u := getUnmarshaler(v); u != nil {
		return u
	}
	if v.CanAddr() {
		if u := getUnmarshaler(v.Addr()); u != nil {
			return u
		}
	}
	if u := getUnmarshaler(reflect.Indirect(v)); u != nil {
		return u
	}
	return nil
}

// AssignStringToValue tries to convert the string to the appropriate type and assign it to the destination variable.
func AssignStringToValue(dst reflect.Value, src string) (err error) {
	if !dst.CanSet() {
		return errValueIsNotAssignable
	}

	v := dst
	var ptr reflect.Value
	// Allocate if pointer is nil
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			ptr = reflect.New(v.Type().Elem())
			v = ptr.Elem()
		} else {
			v = v.Elem()
		}
	}
	if u := findUnmarshaler(v); u != nil {
		return u.UnmarshalText([]byte(src))
	}
	if v.CanInterface() {
		if _, ok := v.Interface().(time.Duration); ok {
			duration, err := time.ParseDuration(src)
			if err == nil {
				v.Set(reflect.ValueOf(duration))
			}
			return err
		}
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(src, 0, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ui, err := strconv.ParseUint(src, 0, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetUint(ui)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(src, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(src)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.String:
		v.SetString(src)
	default:
		err = fmt.Errorf("Unable to convert string \"%s\" to type \"%s\"", src, v.Type().Name())
	}
	if ptr.Kind() == reflect.Ptr {
		dst.Set(ptr)
	}
	return
}

// AssignString tries to convert the string to the appropriate type and assign it to the destination variable.
func AssignString(dst interface{}, src string) error {
	return AssignStringToValue(reflect.Indirect(reflect.ValueOf(dst)), src)
}

// AssignValue tries to convert source to destination. If possible, it converts the string to the destination type.
func AssignValue(dst, src reflect.Value) (err error) {
	dst = reflect.Indirect(dst)
	if !dst.CanSet() {
		return errValueIsNotAssignable
	}
	src = reflect.Indirect(src)

	if src.Kind() == reflect.String {
		return AssignStringToValue(dst, src.String())
	}

	if !src.Type().ConvertibleTo(dst.Type()) {
		return fmt.Errorf("Value of type \"%s\" cannot be converted to type \"%s\"", src.Type().Name(), dst.Type().Name())
	}
	value := src.Convert(dst.Type())
	dst.Set(value)
	return nil
}

// Assign tries to convert source to destination. If possible, it converts the string to the destination type.
func Assign(dst, src interface{}) error {
	return AssignValue(reflect.ValueOf(dst), reflect.ValueOf(src))
}

// ProcessValue is type of callback function for Traverse function.
type ProcessValue func(value reflect.Value, path string, level uint, field *reflect.StructField) error

// Traverse iterates through all the nested elements of the passed variable.
func Traverse(v interface{}, process ProcessValue) error {
	return TraverseValue(reflect.ValueOf(v), process)
}

// TraverseValue iterates through all the nested elements of the passed variable.
func TraverseValue(v reflect.Value, process ProcessValue) error {
	return traverseValue(v, "", 0, nil, process)
}

func addFieldName(path, field string) string {
	if path == "" {
		return field
	}
	return path + "." + field
}

func traverseValue(v reflect.Value, path string, depth uint, field *reflect.StructField, process ProcessValue) error {
	//v = reflect.Indirect(v)
	if err := process(v, path, depth, field); err != nil {
		return err
	}
	depth++

	switch v.Kind() {
	case reflect.Struct:
		structType := v.Type()
		for i := 0; i < structType.NumField(); i++ {
			structField := structType.Field(i)
			fieldValue := v.Field(i)
			if err := traverseValue(fieldValue, addFieldName(path, structField.Name), depth, &structField, process); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		length := v.Len()
		for i := 0; i < length; i++ {
			if err := traverseValue(v.Index(i), path+"["+strconv.Itoa(i)+"]", depth, nil, process); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			keyStr := "[]"
			if key.CanInterface() {
				keyStr = fmt.Sprintf("[%v]", key.Interface())
			}
			if err := traverseValue(v.MapIndex(key), path+keyStr, depth, nil, process); err != nil {
				return err
			}
		}
	case reflect.Ptr:
		if !v.IsNil() {
			if err := traverseValue(v.Elem(), "*("+path+")", depth, nil, process); err != nil {
				return err
			}
		}
	}

	return nil
}

// TraverseFields iterates through all structs's fields of the passed variable.
func TraverseFields(v interface{}, processField ProcessValue) error {
	return TraverseValueFields(reflect.ValueOf(v), processField)
}

// TraverseValueFields iterates through all structs's fields of the passed variable.
func TraverseValueFields(v reflect.Value, processField ProcessValue) error {
	process := func(value reflect.Value, path string, level uint, field *reflect.StructField) error {
		if field != nil {
			return processField(value, path, level, field)
		}
		return nil
	}
	return traverseValue(v, "", 0, nil, process)
}

// Clear variable
func Clear(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}
