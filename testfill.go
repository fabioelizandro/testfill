package testfill

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func Fill[T any](v T) (T, error) {
	var zero T
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	if typ.Kind() != reflect.Struct {
		return zero, fmt.Errorf("testfill: expected struct, got %T", v)
	}

	result := reflect.New(typ).Elem()
	result.Set(val)

	err := fillStruct(result)
	if err != nil {
		return zero, err
	}

	return result.Interface().(T), nil
}

func fillStruct(v reflect.Value) error {
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("testfill: expected struct, got %s", v.Kind())
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get("testfill")

		// Handle nested structs with "fill" tag
		if field.Kind() == reflect.Struct && tag == "fill" {
			err := fillStruct(field)
			if err != nil {
				return fmt.Errorf("testfill: failed to fill nested struct %s: %w", fieldType.Name, err)
			}
			continue
		}

		// Handle pointers to structs with "fill" tag
		if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct && tag == "fill" {
			if field.IsNil() {
				// Create new instance if nil
				newValue := reflect.New(field.Type().Elem())
				field.Set(newValue)
			}
			err := fillStruct(field.Elem())
			if err != nil {
				return fmt.Errorf("testfill: failed to fill nested struct pointer %s: %w", fieldType.Name, err)
			}
			continue
		}

		// Skip fields without testfill tag or with "fill" tag for non-struct types
		if tag == "" || tag == "fill" {
			continue
		}

		if !isZeroValue(field) {
			continue
		}

		err := setFieldValue(field, tag)
		if err != nil {
			return fmt.Errorf("testfill: failed to set field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

func isZeroValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	return v.IsZero()
}

func setFieldValue(field reflect.Value, tag string) error {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(tag, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(val)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(tag, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(val)

	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(tag, 64)
		if err != nil {
			return err
		}
		field.SetFloat(val)

	case reflect.String:
		field.SetString(tag)

	case reflect.Bool:
		val, err := strconv.ParseBool(tag)
		if err != nil {
			return err
		}
		field.SetBool(val)

	case reflect.Slice:
		return setSliceValue(field, tag)

	case reflect.Map:
		return setMapValue(field, tag)

	case reflect.Ptr:
		return setPtrValue(field, tag)

	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			return setTimeValue(field, tag)
		}
		return fmt.Errorf("unsupported struct type %s", field.Type())

	default:
		return fmt.Errorf("unsupported field type %s", field.Kind())
	}

	return nil
}

func setSliceValue(field reflect.Value, tag string) error {
	if field.Type().Elem().Kind() != reflect.String {
		return fmt.Errorf("only string slices are supported")
	}

	parts := strings.Split(tag, ",")
	slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))

	for i, part := range parts {
		slice.Index(i).SetString(strings.TrimSpace(part))
	}

	field.Set(slice)
	return nil
}

func setMapValue(field reflect.Value, tag string) error {
	if field.Type().Key().Kind() != reflect.String || field.Type().Elem().Kind() != reflect.String {
		return fmt.Errorf("only string->string maps are supported")
	}

	m := reflect.MakeMap(field.Type())
	pairs := strings.Split(tag, ",")

	for _, pair := range pairs {
		kv := strings.Split(strings.TrimSpace(pair), ":")
		if len(kv) != 2 {
			return fmt.Errorf("invalid map format: %s", pair)
		}
		key := reflect.ValueOf(strings.TrimSpace(kv[0]))
		value := reflect.ValueOf(strings.TrimSpace(kv[1]))
		m.SetMapIndex(key, value)
	}

	field.Set(m)
	return nil
}

func setPtrValue(field reflect.Value, tag string) error {
	elemType := field.Type().Elem()
	elem := reflect.New(elemType).Elem()

	err := setFieldValue(elem, tag)
	if err != nil {
		return err
	}

	field.Set(elem.Addr())
	return nil
}

func setTimeValue(field reflect.Value, tag string) error {
	t, err := time.Parse(time.RFC3339, tag)
	if err != nil {
		return err
	}
	field.Set(reflect.ValueOf(t))
	return nil
}
