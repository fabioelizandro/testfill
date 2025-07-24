package testfill

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Tag constants
const (
	TagName      = "testfill"
	TagFill      = "fill"
	TagFactory   = "factory:"
	TagUnmarshal = "unmarshal:"
	TagVariant   = "variants:"
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
	ErrJSONUnmarshal        = "failed to unmarshal JSON: %w"
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

// MustFill is like Fill but panics on error.
// Use this when you are certain the struct is valid and want to avoid error handling.
func MustFill[T any](input T) T {
	result, err := Fill(input)
	if err != nil {
		panic(err)
	}

	return result
}

// FillWithVariant populates zero-valued fields in a struct based on testfill tags with a specific variant.
// It takes a struct value and a variant name, returning a copy with fields filled according to their
// variant-specific tags (e.g., testfill_admin) or falling back to default testfill tags.
// Supports nested structs, pointers, slices, maps, and factory functions.
func FillWithVariant[T any](input T, variant string) (T, error) {
	var zero T
	inputValue := reflect.ValueOf(input)
	inputType := reflect.TypeOf(input)

	if inputType.Kind() != reflect.Struct {
		return zero, fmt.Errorf(ErrNotStruct, input)
	}

	// Create a copy to work with
	resultValue := reflect.New(inputType).Elem()
	resultValue.Set(inputValue)

	if err := fillStructWithVariant(resultValue, variant); err != nil {
		return zero, err
	}

	return resultValue.Interface().(T), nil
}

// MustFillWithVariant is like FillWithVariant but panics on error.
// Use this when you are certain the struct is valid and want to avoid error handling.
func MustFillWithVariant[T any](input T, variant string) T {
	result, err := FillWithVariant(input, variant)
	if err != nil {
		panic(err)
	}

	return result
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
	return fillStructWithVariant(structValue, "")
}

func fillStructWithVariant(structValue reflect.Value, variant string) error {
	structType := structValue.Type()
	for i := 0; i < structValue.NumField(); i++ {
		fieldValue := structValue.Field(i)
		fieldType := structType.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Get the appropriate tag value based on variant
		tagValue := getTagValueForVariant(fieldType, variant)

		// Handle nested structs and pointers
		if tagValue == TagFill {
			if err := handleNestedFillWithVariant(fieldValue, fieldType, variant); err != nil {
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

// getTagValueForVariant gets the appropriate tag value based on the variant
// If variant is empty, uses the default "testfill" tag
// If variant is specified, looks for "testfill_<variant>" tag first, falls back to default
func getTagValueForVariant(fieldType reflect.StructField, variant string) string {
	if variant == "" {
		return fieldType.Tag.Get(TagName)
	}

	// Look for variant-specific tag first
	variantTag := TagName + "_" + variant
	if value := fieldType.Tag.Get(variantTag); value != "" {
		return value
	}

	// Fall back to default tag
	return fieldType.Tag.Get(TagName)
}

// =====================================================
// Nested struct handling
// =====================================================

func handleNestedFillWithVariant(field reflect.Value, fieldType reflect.StructField, variant string) error {
	switch field.Kind() {
	case reflect.Struct:
		if err := fillStructWithVariant(field, variant); err != nil {
			return fmt.Errorf(ErrNestedStruct, fieldType.Name, err)
		}
	case reflect.Ptr:
		if field.Type().Elem().Kind() == reflect.Struct {
			if field.IsNil() {
				// Create new instance if nil
				newValue := reflect.New(field.Type().Elem())
				field.Set(newValue)
			}
			if err := fillStructWithVariant(field.Elem(), variant); err != nil {
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
	// Handle JSON unmarshal
	if strings.HasPrefix(tag, TagUnmarshal) {
		jsonData := strings.TrimPrefix(tag, TagUnmarshal)
		return unmarshalJSON(field, jsonData)
	}

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

	// Support "variants:name1,name2,name3" syntax for struct slices with different field values
	if strings.HasPrefix(tag, TagVariant) {
		variantStr := strings.TrimPrefix(tag, TagVariant)
		variants := strings.Split(variantStr, ",")

		// Clean up variant names
		for i, variant := range variants {
			variants[i] = strings.TrimSpace(variant)
		}

		slice := reflect.MakeSlice(field.Type(), len(variants), len(variants))
		for i, variant := range variants {
			elemValue := reflect.New(elemType).Elem()
			if err := fillStructWithVariant(elemValue, variant); err != nil {
				return fmt.Errorf("failed to fill slice element %d with variant %s: %w", i, variant, err)
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

	// Check if this is a variants syntax
	if strings.HasPrefix(tag, "variants:") {
		return setStructMapWithVariants(field, tag, valueType)
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
			// Create and fill a new struct instance with default variant
			structValue := reflect.New(valueType).Elem()
			if err := fillStructWithVariant(structValue, ""); err != nil {
				return fmt.Errorf("failed to fill map value for key %s: %w", keyStr, err)
			}
			m.SetMapIndex(keyValue, structValue)
		} else {
			// Assume valueStr is a variant name
			structValue := reflect.New(valueType).Elem()
			if err := fillStructWithVariant(structValue, valueStr); err != nil {
				return fmt.Errorf("failed to fill map value for key %s with variant %s: %w", keyStr, valueStr, err)
			}
			m.SetMapIndex(keyValue, structValue)
		}
	}

	field.Set(m)
	return nil
}

func setStructMapWithVariants(field reflect.Value, tag string, valueType reflect.Type) error {
	// Extract variants from "variants:key1=variant1,key2=variant2,..." syntax
	variantStr := strings.TrimPrefix(tag, "variants:")
	items := strings.Split(variantStr, ",")

	// Clean up items
	for i, item := range items {
		items[i] = strings.TrimSpace(item)
	}

	m := reflect.MakeMap(field.Type())

	for _, item := range items {
		// All items must use key=variant syntax
		kv := strings.Split(item, "=")
		if len(kv) != 2 {
			return fmt.Errorf("invalid key=variant format: %s (expected format: key=variant)", item)
		}

		keyStr := strings.TrimSpace(kv[0])
		variant := strings.TrimSpace(kv[1])

		keyValue := reflect.ValueOf(keyStr)

		// Create and fill struct with the specified variant
		structValue := reflect.New(valueType).Elem()
		if err := fillStructWithVariant(structValue, variant); err != nil {
			return fmt.Errorf("failed to fill map value for key %s with variant %s: %w", keyStr, variant, err)
		}
		m.SetMapIndex(keyValue, structValue)
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

// =====================================================
// JSON unmarshal support
// =====================================================

func unmarshalJSON(field reflect.Value, jsonData string) error {
	if field.Kind() == reflect.Ptr {
		if jsonData == "null" {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}

		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}

		// Unmarshal into the pointed value
		return unmarshalJSONValue(field.Interface(), jsonData)
	}

	// For non-pointer types, we need to unmarshal into the address
	if field.CanAddr() {
		return unmarshalJSONValue(field.Addr().Interface(), jsonData)
	}

	// If we can't get the address, create a new value, unmarshal, and set
	newValue := reflect.New(field.Type())
	if err := unmarshalJSONValue(newValue.Interface(), jsonData); err != nil {
		return err
	}
	field.Set(newValue.Elem())
	return nil
}

func unmarshalJSONValue(target interface{}, jsonData string) error {
	if err := json.Unmarshal([]byte(jsonData), target); err != nil {
		return fmt.Errorf(ErrJSONUnmarshal, err)
	}
	return nil
}
