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

		err := setFieldValue(field, fieldType, tag)
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

func setFieldValue(field reflect.Value, fieldType reflect.StructField, tag string) error {
	// Handle factory functions
	if strings.HasPrefix(tag, "factory:") {
		factoryTag := strings.TrimPrefix(tag, "factory:")
		return callFactoryFunction(field, fieldType, factoryTag)
	}
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
		return setPtrValue(field, fieldType, tag)

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

func setPtrValue(field reflect.Value, fieldType reflect.StructField, tag string) error {
	elemType := field.Type().Elem()
	elem := reflect.New(elemType).Elem()

	err := setFieldValue(elem, fieldType, tag)
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

func callFactoryFunction(field reflect.Value, fieldType reflect.StructField, factoryTag string) error {
	// Parse factory name and arguments from tag
	// Format: "FunctionName" or "FunctionName:arg1:arg2..."
	parts := strings.Split(factoryTag, ":")
	factoryName := parts[0]
	args := parts[1:]

	funcValue := reflect.ValueOf(getFactoryFunction(factoryName))
	if !funcValue.IsValid() {
		return fmt.Errorf("factory function %s not found", factoryName)
	}

	funcType := funcValue.Type()

	// Validate argument count
	if len(args) != funcType.NumIn() {
		return fmt.Errorf("factory function %s expects %d arguments, got %d",
			factoryName, funcType.NumIn(), len(args))
	}

	// Prepare arguments
	callArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		paramType := funcType.In(i)

		// Convert string argument to the expected parameter type
		argValue, err := convertStringToType(arg, paramType)
		if err != nil {
			return fmt.Errorf("factory function %s argument %d: %w", factoryName, i, err)
		}
		callArgs[i] = argValue
	}

	// Call the factory function
	results := funcValue.Call(callArgs)
	if len(results) != 1 {
		return fmt.Errorf("factory function %s must return exactly one value", factoryName)
	}

	result := results[0]
	if !result.Type().AssignableTo(field.Type()) {
		return fmt.Errorf("factory function %s returns %s, but field expects %s",
			factoryName, result.Type(), field.Type())
	}

	field.Set(result)
	return nil
}

// This is a simplified registry for factory functions
// In a real implementation, you might want a more sophisticated approach
var factoryRegistry = make(map[string]interface{})

func getFactoryFunction(name string) interface{} {
	if fn, exists := factoryRegistry[name]; exists {
		return fn
	}

	// Factory functions must be registered before use

	return nil
}

// This would be a public function to register factory functions
func RegisterFactory(name string, fn interface{}) {
	factoryRegistry[name] = fn
}

// convertStringToType converts a string argument to the expected parameter type
func convertStringToType(arg string, targetType reflect.Type) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.String:
		return reflect.ValueOf(arg), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert %q to %s: %w", arg, targetType.Kind(), err)
		}
		return reflect.ValueOf(val).Convert(targetType), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(arg, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert %q to %s: %w", arg, targetType.Kind(), err)
		}
		return reflect.ValueOf(val).Convert(targetType), nil
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert %q to %s: %w", arg, targetType.Kind(), err)
		}
		return reflect.ValueOf(val).Convert(targetType), nil
	case reflect.Bool:
		val, err := strconv.ParseBool(arg)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert %q to %s: %w", arg, targetType.Kind(), err)
		}
		return reflect.ValueOf(val), nil
	default:
		return reflect.Value{}, fmt.Errorf("unsupported parameter type %s for factory function arguments", targetType.Kind())
	}
}
