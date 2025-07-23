package testfill

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Tag constants
const (
	TagName    = "testfill"
	TagFill    = "fill"
	TagFactory = "factory:"
)

// Error messages
const (
	ErrNotStruct            = "testfill: expected struct, got %T"
	ErrNestedStruct         = "testfill: failed to fill nested struct %s: %w"
	ErrNestedStructPtr      = "testfill: failed to fill nested struct pointer %s: %w"
	ErrSetField             = "testfill: failed to set field %s: %w"
	ErrUnsupportedStruct    = "unsupported struct type %s"
	ErrUnsupportedField     = "unsupported field type %s"
	ErrUnsupportedSliceType = "unsupported slice element type %s"
	ErrUnsupportedMapType   = "unsupported map type %s -> %s"
	ErrInvalidMapFormat     = "invalid map format: %s"
	ErrFactoryNotFound      = "factory function %s not found"
	ErrFactoryArgCount      = "factory function %s expects %d arguments, got %d"
	ErrFactoryPanic         = "factory function panicked: %v"
	ErrFactoryReturnCount   = "factory function %s must return exactly one value"
	ErrFactoryReturnType    = "factory function %s returns %s, but field expects %s"
	ErrFactoryArgConvert    = "factory function %s argument %d: %w"
	ErrStringConvert        = "cannot convert %q to %s: %w"
	ErrUnsupportedParam     = "unsupported parameter type %s for factory function arguments"
)

// =====================================================
// Main API Functions
// =====================================================

// Fill populates zero-valued fields in a struct based on testfill tags.
// It takes a struct value and returns a copy with fields filled according to their tags.
// Supports nested structs, pointers, slices, maps, and factory functions.
func Fill[T any](input T) (T, error) {
	var zero T
	inputValue := reflect.ValueOf(input)
	inputType := reflect.TypeOf(input)

	if inputType.Kind() != reflect.Struct {
		return zero, fmt.Errorf(ErrNotStruct, input)
	}

	// Create a copy to work with
	resultValue := reflect.New(inputType).Elem()
	resultValue.Set(inputValue)

	if err := fillStruct(resultValue); err != nil {
		return zero, err
	}

	return resultValue.Interface().(T), nil
}

// RegisterFactory registers a factory function that can be called from struct tags.
// The function must return exactly one value that matches the field type.
// Factory functions can accept string arguments that will be converted to the appropriate types.
//
// Example:
//	// Register a factory function
//	testfill.RegisterFactory("uuid", func() string { return "test-uuid-123" })
//	
//	// Use in struct tag
//	type User struct {
//		ID string `testfill:"factory:uuid"`
//	}
func RegisterFactory(name string, fn interface{}) {
	factoryRegistry[name] = fn
}

// =====================================================
// Core struct filling logic
// =====================================================

func fillStruct(structValue reflect.Value) error {
	structType := structValue.Type()
	for i := 0; i < structValue.NumField(); i++ {
		fieldValue := structValue.Field(i)
		fieldType := structType.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		tagValue := fieldType.Tag.Get(TagName)

		// Handle nested structs and pointers
		if tagValue == TagFill {
			if err := handleNestedFill(fieldValue, fieldType); err != nil {
				return err
			}
			continue
		}

		// Skip fields without testfill tag
		if tagValue == "" {
			continue
		}

		// Skip non-zero fields
		if !isZeroValue(fieldValue) {
			continue
		}

		if err := setFieldValue(fieldValue, fieldType, tagValue); err != nil {
			return fmt.Errorf(ErrSetField, fieldType.Name, err)
		}
	}

	return nil
}

// =====================================================
// Reflection utility functions
// =====================================================

func isZeroValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	return v.IsZero()
}

// =====================================================
// Nested struct handling
// =====================================================

func handleNestedFill(field reflect.Value, fieldType reflect.StructField) error {
	switch field.Kind() {
	case reflect.Struct:
		if err := fillStruct(field); err != nil {
			return fmt.Errorf(ErrNestedStruct, fieldType.Name, err)
		}
	case reflect.Ptr:
		if field.Type().Elem().Kind() == reflect.Struct {
			if field.IsNil() {
				// Create new instance if nil
				newValue := reflect.New(field.Type().Elem())
				field.Set(newValue)
			}
			if err := fillStruct(field.Elem()); err != nil {
				return fmt.Errorf(ErrNestedStructPtr, fieldType.Name, err)
			}
		}
	}
	return nil
}

// =====================================================
// Field value setting
// =====================================================

func setFieldValue(field reflect.Value, _ reflect.StructField, tag string) error {
	// Handle factory functions
	if strings.HasPrefix(tag, TagFactory) {
		factoryTag := strings.TrimPrefix(tag, TagFactory)
		return callFactoryFunction(field, factoryTag)
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
		return setPrimitiveValue(field, tag)
	case reflect.Slice:
		return setSliceValue(field, tag)
	case reflect.Map:
		return setMapValue(field, tag)
	case reflect.Ptr:
		return setPtrValue(field, tag)
	case reflect.Struct:
		return setStructValue(field, tag)
	default:
		return fmt.Errorf(ErrUnsupportedField, field.Kind())
	}
}

func setSliceValue(field reflect.Value, tag string) error {
	elemType := field.Type().Elem()

	// Handle struct slices with special "fill:count" syntax
	if elemType.Kind() == reflect.Struct {
		return setStructSliceValue(field, tag, elemType)
	}

	// Handle primitive slices
	parts := strings.Split(tag, ",")
	slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))

	for i, part := range parts {
		elemValue, err := convertStringToType(strings.TrimSpace(part), elemType)
		if err != nil {
			return fmt.Errorf(ErrUnsupportedSliceType, elemType.Kind())
		}
		slice.Index(i).Set(elemValue)
	}

	field.Set(slice)
	return nil
}

func setStructSliceValue(field reflect.Value, tag string, elemType reflect.Type) error {
	// Support "fill:count" syntax for struct slices
	if strings.HasPrefix(tag, "fill:") {
		countStr := strings.TrimPrefix(tag, "fill:")
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return fmt.Errorf("invalid slice count format: %s", tag)
		}

		slice := reflect.MakeSlice(field.Type(), count, count)
		for i := 0; i < count; i++ {
			elemValue := reflect.New(elemType).Elem()
			if err := fillStruct(elemValue); err != nil {
				return fmt.Errorf("failed to fill slice element %d: %w", i, err)
			}
			slice.Index(i).Set(elemValue)
		}
		field.Set(slice)
		return nil
	}

	return fmt.Errorf(ErrUnsupportedSliceType, elemType.Kind())
}

func setMapValue(field reflect.Value, tag string) error {
	keyType := field.Type().Key()
	valueType := field.Type().Elem()

	// Handle struct value maps with special "key:fill" syntax
	if valueType.Kind() == reflect.Struct {
		return setStructMapValue(field, tag, keyType, valueType)
	}

	// Handle primitive maps
	m := reflect.MakeMap(field.Type())
	pairs := strings.Split(tag, ",")

	for _, pair := range pairs {
		kv := strings.Split(strings.TrimSpace(pair), ":")
		if len(kv) != 2 {
			return fmt.Errorf(ErrInvalidMapFormat, pair)
		}

		keyValue, err := convertStringToType(strings.TrimSpace(kv[0]), keyType)
		if err != nil {
			return fmt.Errorf(ErrUnsupportedMapType, keyType.Kind(), valueType.Kind())
		}

		valueValue, err := convertStringToType(strings.TrimSpace(kv[1]), valueType)
		if err != nil {
			return fmt.Errorf(ErrUnsupportedMapType, keyType.Kind(), valueType.Kind())
		}

		m.SetMapIndex(keyValue, valueValue)
	}

	field.Set(m)
	return nil
}

func setStructMapValue(field reflect.Value, tag string, keyType, valueType reflect.Type) error {
	// Only support string keys for struct value maps
	if keyType.Kind() != reflect.String {
		return fmt.Errorf(ErrUnsupportedMapType, keyType.Kind(), valueType.Kind())
	}

	m := reflect.MakeMap(field.Type())
	pairs := strings.Split(tag, ",")

	for _, pair := range pairs {
		kv := strings.Split(strings.TrimSpace(pair), ":")
		if len(kv) != 2 {
			return fmt.Errorf(ErrInvalidMapFormat, pair)
		}

		keyStr := strings.TrimSpace(kv[0])
		valueStr := strings.TrimSpace(kv[1])

		keyValue := reflect.ValueOf(keyStr)

		if valueStr == "fill" {
			// Create and fill a new struct instance
			structValue := reflect.New(valueType).Elem()
			if err := fillStruct(structValue); err != nil {
				return fmt.Errorf("failed to fill map value for key %s: %w", keyStr, err)
			}
			m.SetMapIndex(keyValue, structValue)
		} else {
			return fmt.Errorf("struct map values must use 'fill' syntax, got: %s", valueStr)
		}
	}

	field.Set(m)
	return nil
}

func setPtrValue(field reflect.Value, tag string) error {
	elemType := field.Type().Elem()
	elem := reflect.New(elemType).Elem()

	// Create a dummy StructField for recursive call
	dummyField := reflect.StructField{Type: elemType}
	err := setFieldValue(elem, dummyField, tag)
	if err != nil {
		return err
	}

	field.Set(elem.Addr())
	return nil
}

// setPrimitiveValue handles all primitive types (int, uint, float, string, bool)
func setPrimitiveValue(field reflect.Value, tag string) error {
	convertedValue, err := convertStringToType(tag, field.Type())
	if err != nil {
		return err
	}
	field.Set(convertedValue)
	return nil
}

func setStructValue(field reflect.Value, tag string) error {
	if field.Type() == reflect.TypeOf(time.Time{}) {
		return setTimeValue(field, tag)
	}
	return fmt.Errorf(ErrUnsupportedStruct, field.Type())
}

func setTimeValue(field reflect.Value, tag string) error {
	t, err := time.Parse(time.RFC3339, tag)
	if err != nil {
		return err
	}
	field.Set(reflect.ValueOf(t))
	return nil
}

func callFactoryFunction(field reflect.Value, factoryTag string) (err error) {
	// Recover from panics in factory functions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(ErrFactoryPanic, r)
		}
	}()

	factoryName, args := parseFactoryTag(factoryTag)
	funcValue, funcType, err := getAndValidateFactoryFunction(factoryName)
	if err != nil {
		return err
	}

	callArgs, err := prepareFactoryArgs(args, funcType, factoryName)
	if err != nil {
		return err
	}

	result, err := callAndValidateFactory(funcValue, callArgs, factoryName, field.Type())
	if err != nil {
		return err
	}

	field.Set(result)
	return nil
}

// =====================================================
// Factory function system
// =====================================================

func parseFactoryTag(factoryTag string) (string, []string) {
	// Parse factory name and arguments from tag
	// Format: "FunctionName" or "FunctionName:arg1:arg2..."
	parts := strings.Split(factoryTag, ":")
	return parts[0], parts[1:]
}

func getAndValidateFactoryFunction(factoryName string) (reflect.Value, reflect.Type, error) {
	funcValue := reflect.ValueOf(getFactoryFunction(factoryName))
	if !funcValue.IsValid() {
		return reflect.Value{}, nil, fmt.Errorf(ErrFactoryNotFound, factoryName)
	}
	return funcValue, funcValue.Type(), nil
}

func prepareFactoryArgs(args []string, funcType reflect.Type, factoryName string) ([]reflect.Value, error) {
	// Validate argument count
	if len(args) != funcType.NumIn() {
		return nil, fmt.Errorf(ErrFactoryArgCount, factoryName, funcType.NumIn(), len(args))
	}

	// Prepare arguments
	callArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		paramType := funcType.In(i)
		argValue, err := convertStringToType(arg, paramType)
		if err != nil {
			return nil, fmt.Errorf(ErrFactoryArgConvert, factoryName, i, err)
		}
		callArgs[i] = argValue
	}
	return callArgs, nil
}

func callAndValidateFactory(funcValue reflect.Value, callArgs []reflect.Value, factoryName string, fieldType reflect.Type) (reflect.Value, error) {
	// Call the factory function
	results := funcValue.Call(callArgs)
	if len(results) != 1 {
		return reflect.Value{}, fmt.Errorf(ErrFactoryReturnCount, factoryName)
	}

	result := results[0]
	if !result.Type().AssignableTo(fieldType) {
		return reflect.Value{}, fmt.Errorf(ErrFactoryReturnType, factoryName, result.Type(), fieldType)
	}
	return result, nil
}

// =====================================================
// Factory registry and public API
// =====================================================

// Factory registry
var factoryRegistry = make(map[string]interface{})

func getFactoryFunction(name string) interface{} {
	if fn, exists := factoryRegistry[name]; exists {
		return fn
	}

	// Factory functions must be registered before use

	return nil
}

// =====================================================
// Type conversion utilities
// ==============================================

type typeConverter func(string) (interface{}, error)

var typeConverters = map[reflect.Kind]typeConverter{
	reflect.String:  func(s string) (interface{}, error) { return s, nil },
	reflect.Bool:    func(s string) (interface{}, error) { return strconv.ParseBool(s) },
	reflect.Int:     func(s string) (interface{}, error) { return strconv.ParseInt(s, 10, 64) },
	reflect.Int8:    func(s string) (interface{}, error) { return strconv.ParseInt(s, 10, 8) },
	reflect.Int16:   func(s string) (interface{}, error) { return strconv.ParseInt(s, 10, 16) },
	reflect.Int32:   func(s string) (interface{}, error) { return strconv.ParseInt(s, 10, 32) },
	reflect.Int64:   func(s string) (interface{}, error) { return strconv.ParseInt(s, 10, 64) },
	reflect.Uint:    func(s string) (interface{}, error) { return strconv.ParseUint(s, 10, 64) },
	reflect.Uint8:   func(s string) (interface{}, error) { return strconv.ParseUint(s, 10, 8) },
	reflect.Uint16:  func(s string) (interface{}, error) { return strconv.ParseUint(s, 10, 16) },
	reflect.Uint32:  func(s string) (interface{}, error) { return strconv.ParseUint(s, 10, 32) },
	reflect.Uint64:  func(s string) (interface{}, error) { return strconv.ParseUint(s, 10, 64) },
	reflect.Float32: func(s string) (interface{}, error) { return strconv.ParseFloat(s, 32) },
	reflect.Float64: func(s string) (interface{}, error) { return strconv.ParseFloat(s, 64) },
}

func convertStringToType(arg string, targetType reflect.Type) (reflect.Value, error) {
	converter, exists := typeConverters[targetType.Kind()]
	if !exists {
		return reflect.Value{}, fmt.Errorf(ErrUnsupportedParam, targetType.Kind())
	}

	val, err := converter(arg)
	if err != nil {
		return reflect.Value{}, fmt.Errorf(ErrStringConvert, arg, targetType.Kind(), err)
	}

	return reflect.ValueOf(val).Convert(targetType), nil
}
