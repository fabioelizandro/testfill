package testfill_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/fabioelizandro/testfill"
	"github.com/stretchr/testify/require"
)

type Foo struct {
	NestedStructWithFillTag  Bar `testfill:"fill"`
	NestedStructWithoutTag   Bar
	NestedPointerWithFillTag *Bar `testfill:"fill"`
	NestedPointerWithoutTag  *Bar
	DeeplyNestedWithFillTag  Baz `testfill:"fill"`
	DeeplyNestedWithoutTag   Baz
	CustomVOMultiArgs        CustomVO `testfill:"factory:NewCustomVOMultiArgs:prefix:42:suffix"`
}

type Bar struct {
	Integer int    `testfill:"42"`
	String  string `testfill:"Olivie Smith"`
}

type Baz struct {
	Name         string `testfill:"Deep Nested"`
	Value        int    `testfill:"100"`
	NestedBar    Bar    `testfill:"fill"`
	NonFilledBar Bar
}

type CustomVO struct {
	privateField string
}

func TestTestfill(t *testing.T) {
	// Register factory with no arguments
	testfill.RegisterFactory("NewCustomVO", func() CustomVO {
		return CustomVO{privateField: "factory default"}
	})

	// Register factory with arguments
	testfill.RegisterFactory("NewCustomVOWithArg", func(arg string) CustomVO {
		return CustomVO{privateField: arg}
	})

	// Register factory with multiple arguments
	testfill.RegisterFactory("NewCustomVOMultiArgs", func(prefix string, number int, suffix string) CustomVO {
		return CustomVO{privateField: fmt.Sprintf("%s-%d-%s", prefix, number, suffix)}
	})

	// Register time factories
	testfill.RegisterFactory("ParseDate", func(dateStr string) time.Time {
		// Parse date in YYYY-MM-DD format and set time to midnight UTC
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			panic(fmt.Sprintf("invalid date format (expected YYYY-MM-DD): %s", dateStr))
		}
		return t.UTC()
	})

	// Register a factory that panics for testing
	testfill.RegisterFactory("PanicFactory", func() CustomVO {
		panic("this factory always panics")
	})

	t.Run("integers", func(t *testing.T) {
		t.Run("int fills default value", func(t *testing.T) {
			type IntTest struct {
				Value int `testfill:"42"`
			}

			result, err := testfill.Fill(IntTest{})
			require.NoError(t, err)

			require.Equal(t, 42, result.Value)
		})

		t.Run("int8 fills default value", func(t *testing.T) {
			type Int8Test struct {
				Value int8 `testfill:"127"`
			}

			result, err := testfill.Fill(Int8Test{})
			require.NoError(t, err)

			require.Equal(t, int8(127), result.Value)
		})

		t.Run("int16 fills default value", func(t *testing.T) {
			type Int16Test struct {
				Value int16 `testfill:"32767"`
			}

			result, err := testfill.Fill(Int16Test{})
			require.NoError(t, err)

			require.Equal(t, int16(32767), result.Value)
		})

		t.Run("int32 fills default value", func(t *testing.T) {
			type Int32Test struct {
				Value int32 `testfill:"2147483647"`
			}

			result, err := testfill.Fill(Int32Test{})
			require.NoError(t, err)

			require.Equal(t, int32(2147483647), result.Value)
		})

		t.Run("int64 fills default value", func(t *testing.T) {
			type Int64Test struct {
				Value int64 `testfill:"9223372036854775807"`
			}

			result, err := testfill.Fill(Int64Test{})
			require.NoError(t, err)

			require.Equal(t, int64(9223372036854775807), result.Value)
		})

		t.Run("does not fill when value is already filled", func(t *testing.T) {
			type IntTest struct {
				Value int `testfill:"42"`
			}

			result, err := testfill.Fill(IntTest{Value: 21})
			require.NoError(t, err)

			require.Equal(t, 21, result.Value)
		})

		t.Run("invalid int tag", func(t *testing.T) {
			type InvalidInt struct {
				Value int `testfill:"not_a_number"`
			}

			result, err := testfill.Fill(InvalidInt{})

			expectedError := "testfill: failed to set field Value: cannot convert \"not_a_number\" to int: strconv.ParseInt: parsing \"not_a_number\": invalid syntax"
			require.EqualError(t, err, expectedError)
			require.Equal(t, InvalidInt{}, result)
		})

		t.Run("pointer to int fills value", func(t *testing.T) {
			type PointerTest struct {
				Value *int `testfill:"42"`
			}

			result, err := testfill.Fill(PointerTest{})
			require.NoError(t, err)

			require.NotNil(t, result.Value)
			require.Equal(t, 42, *result.Value)
		})

		t.Run("pointer to invalid int", func(t *testing.T) {
			type PtrErrorStruct struct {
				IntPtr *int `testfill:"not_a_number"`
			}

			result, err := testfill.Fill(PtrErrorStruct{})

			expectedError := "testfill: failed to set field IntPtr: cannot convert \"not_a_number\" to int: strconv.ParseInt: parsing \"not_a_number\": invalid syntax"
			require.EqualError(t, err, expectedError)
			require.Equal(t, PtrErrorStruct{}, result)
		})
	})

	t.Run("unsigned integers", func(t *testing.T) {
		t.Run("uint fills default value", func(t *testing.T) {
			type UintTest struct {
				Value uint `testfill:"42"`
			}

			result, err := testfill.Fill(UintTest{})
			require.NoError(t, err)

			require.Equal(t, uint(42), result.Value)
		})

		t.Run("uint8 fills default value", func(t *testing.T) {
			type Uint8Test struct {
				Value uint8 `testfill:"255"`
			}

			result, err := testfill.Fill(Uint8Test{})
			require.NoError(t, err)

			require.Equal(t, uint8(255), result.Value)
		})

		t.Run("uint16 fills default value", func(t *testing.T) {
			type Uint16Test struct {
				Value uint16 `testfill:"65535"`
			}

			result, err := testfill.Fill(Uint16Test{})
			require.NoError(t, err)

			require.Equal(t, uint16(65535), result.Value)
		})

		t.Run("uint32 fills default value", func(t *testing.T) {
			type Uint32Test struct {
				Value uint32 `testfill:"4294967295"`
			}

			result, err := testfill.Fill(Uint32Test{})
			require.NoError(t, err)

			require.Equal(t, uint32(4294967295), result.Value)
		})

		t.Run("uint64 fills default value", func(t *testing.T) {
			type Uint64Test struct {
				Value uint64 `testfill:"18446744073709551615"`
			}

			result, err := testfill.Fill(Uint64Test{})
			require.NoError(t, err)

			require.Equal(t, uint64(18446744073709551615), result.Value)
		})

		t.Run("invalid uint tag", func(t *testing.T) {
			type InvalidUint struct {
				Value uint `testfill:"not_a_number"`
			}

			result, err := testfill.Fill(InvalidUint{})

			expectedError := "testfill: failed to set field Value: cannot convert \"not_a_number\" to uint: strconv.ParseUint: parsing \"not_a_number\": invalid syntax"
			require.EqualError(t, err, expectedError)
			require.Equal(t, InvalidUint{}, result)
		})

		t.Run("pointer to uint fills value", func(t *testing.T) {
			type PointerTest struct {
				Value *uint `testfill:"42"`
			}

			result, err := testfill.Fill(PointerTest{})
			require.NoError(t, err)

			require.NotNil(t, result.Value)
			require.Equal(t, uint(42), *result.Value)
		})
	})

	t.Run("string", func(t *testing.T) {
		t.Run("fills default value", func(t *testing.T) {
			type StringTest struct {
				Value string `testfill:"John Doe"`
			}

			result, err := testfill.Fill(StringTest{})
			require.NoError(t, err)

			require.Equal(t, "John Doe", result.Value)
		})

		t.Run("does not fill when value is already filled", func(t *testing.T) {
			type StringTest struct {
				Value string `testfill:"John Doe"`
			}

			result, err := testfill.Fill(StringTest{Value: "Mary"})
			require.NoError(t, err)

			require.Equal(t, "Mary", result.Value)
		})

		t.Run("pointer to string fills value", func(t *testing.T) {
			type PointerTest struct {
				Value *string `testfill:"hello"`
			}

			result, err := testfill.Fill(PointerTest{})
			require.NoError(t, err)

			require.NotNil(t, result.Value)
			require.Equal(t, "hello", *result.Value)
		})
	})

	t.Run("boolean", func(t *testing.T) {
		t.Run("fills default value", func(t *testing.T) {
			type BoolTest struct {
				Value bool `testfill:"true"`
			}

			result, err := testfill.Fill(BoolTest{})
			require.NoError(t, err)

			require.Equal(t, true, result.Value)
		})

		t.Run("does not fill when value is already filled", func(t *testing.T) {
			type BoolTest struct {
				Value bool `testfill:"false"`
			}

			result, err := testfill.Fill(BoolTest{Value: true})
			require.NoError(t, err)

			require.Equal(t, true, result.Value)
		})

		t.Run("invalid bool tag", func(t *testing.T) {
			type InvalidBool struct {
				Value bool `testfill:"not_a_bool"`
			}

			result, err := testfill.Fill(InvalidBool{})

			expectedError := "testfill: failed to set field Value: cannot convert \"not_a_bool\" to bool: strconv.ParseBool: parsing \"not_a_bool\": invalid syntax"
			require.EqualError(t, err, expectedError)
			require.Equal(t, InvalidBool{}, result)
		})

		t.Run("pointer to bool fills value", func(t *testing.T) {
			type PointerTest struct {
				Value *bool `testfill:"true"`
			}

			result, err := testfill.Fill(PointerTest{})
			require.NoError(t, err)

			require.NotNil(t, result.Value)
			require.Equal(t, true, *result.Value)
		})
	})

	t.Run("float", func(t *testing.T) {
		t.Run("float32 fills default value", func(t *testing.T) {
			type Float32Test struct {
				Value float32 `testfill:"99.99"`
			}

			result, err := testfill.Fill(Float32Test{})
			require.NoError(t, err)

			require.Equal(t, float32(99.99), result.Value)
		})

		t.Run("float64 fills default value", func(t *testing.T) {
			type Float64Test struct {
				Value float64 `testfill:"99.99"`
			}

			result, err := testfill.Fill(Float64Test{})
			require.NoError(t, err)

			require.Equal(t, 99.99, result.Value)
		})

		t.Run("does not fill when value is already filled", func(t *testing.T) {
			type FloatTest struct {
				Value float64 `testfill:"99.99"`
			}

			result, err := testfill.Fill(FloatTest{Value: 123.45})
			require.NoError(t, err)

			require.Equal(t, 123.45, result.Value)
		})

		t.Run("invalid float tag", func(t *testing.T) {
			type InvalidFloat struct {
				Value float64 `testfill:"not_a_float"`
			}

			result, err := testfill.Fill(InvalidFloat{})

			expectedError := "testfill: failed to set field Value: cannot convert \"not_a_float\" to float64: strconv.ParseFloat: parsing \"not_a_float\": invalid syntax"
			require.EqualError(t, err, expectedError)
			require.Equal(t, InvalidFloat{}, result)
		})

		t.Run("pointer to float fills value", func(t *testing.T) {
			type PointerTest struct {
				Value *float64 `testfill:"99.99"`
			}

			result, err := testfill.Fill(PointerTest{})
			require.NoError(t, err)

			require.NotNil(t, result.Value)
			require.Equal(t, 99.99, *result.Value)
		})

		t.Run("pointer to invalid float", func(t *testing.T) {
			type PtrErrorStruct struct {
				FloatPtr *float64 `testfill:"not_a_float"`
			}

			result, err := testfill.Fill(PtrErrorStruct{})

			expectedError := "testfill: failed to set field FloatPtr: cannot convert \"not_a_float\" to float64: strconv.ParseFloat: parsing \"not_a_float\": invalid syntax"
			require.EqualError(t, err, expectedError)
			require.Equal(t, PtrErrorStruct{}, result)
		})
	})

	t.Run("time", func(t *testing.T) {
		t.Run("fills default value", func(t *testing.T) {
			type TimeTest struct {
				Value time.Time `testfill:"2023-01-15T10:30:00Z"`
			}

			result, err := testfill.Fill(TimeTest{})
			require.NoError(t, err)

			expected, _ := time.Parse(time.RFC3339, "2023-01-15T10:30:00Z")
			require.Equal(t, expected, result.Value)
		})

		t.Run("does not fill when value is already filled", func(t *testing.T) {
			type TimeTest struct {
				Value time.Time `testfill:"2023-01-15T10:30:00Z"`
			}

			customTime, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
			result, err := testfill.Fill(TimeTest{Value: customTime})
			require.NoError(t, err)

			require.Equal(t, customTime, result.Value)
		})

		t.Run("invalid RFC3339 format", func(t *testing.T) {
			type InvalidTime struct {
				Value time.Time `testfill:"2023-13-45T25:70:99Z"`
			}

			result, err := testfill.Fill(InvalidTime{})

			expectedError := "testfill: failed to set field Value: parsing time \"2023-13-45T25:70:99Z\": month out of range"
			require.EqualError(t, err, expectedError)
			require.Equal(t, InvalidTime{}, result)
		})

		t.Run("non-RFC3339 format", func(t *testing.T) {
			type InvalidTime struct {
				Value time.Time `testfill:"2023-01-15"`
			}

			result, err := testfill.Fill(InvalidTime{})

			expectedError := "testfill: failed to set field Value: parsing time \"2023-01-15\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"\" as \"T\""
			require.EqualError(t, err, expectedError)
			require.Equal(t, InvalidTime{}, result)
		})
	})

	t.Run("array", func(t *testing.T) {
		t.Run("array of string", func(t *testing.T) {
			t.Run("fills default value", func(t *testing.T) {
				type ArrayTest struct {
					Value []string `testfill:"tag1,tag2,tag3"`
				}

				result, err := testfill.Fill(ArrayTest{})
				require.NoError(t, err)

				require.Equal(t, []string{"tag1", "tag2", "tag3"}, result.Value)
			})

			t.Run("does not fill when value is already filled", func(t *testing.T) {
				type ArrayTest struct {
					Value []string `testfill:"tag1,tag2,tag3"`
				}

				custom := []string{"custom1", "custom2"}
				result, err := testfill.Fill(ArrayTest{Value: custom})
				require.NoError(t, err)

				require.Equal(t, custom, result.Value)
			})

			t.Run("int slice", func(t *testing.T) {
				type IntSliceTest struct {
					Value []int `testfill:"1,2,3,42"`
				}

				result, err := testfill.Fill(IntSliceTest{})
				require.NoError(t, err)

				require.Equal(t, []int{1, 2, 3, 42}, result.Value)
			})

			t.Run("float slice", func(t *testing.T) {
				type FloatSliceTest struct {
					Value []float64 `testfill:"1.1,2.5,3.14"`
				}

				result, err := testfill.Fill(FloatSliceTest{})
				require.NoError(t, err)

				require.Equal(t, []float64{1.1, 2.5, 3.14}, result.Value)
			})

			t.Run("bool slice", func(t *testing.T) {
				type BoolSliceTest struct {
					Value []bool `testfill:"true,false,true"`
				}

				result, err := testfill.Fill(BoolSliceTest{})
				require.NoError(t, err)

				require.Equal(t, []bool{true, false, true}, result.Value)
			})

			t.Run("struct slice with fill syntax", func(t *testing.T) {
				type StructSliceTest struct {
					Value []Bar `testfill:"fill:2"`
				}

				result, err := testfill.Fill(StructSliceTest{})
				require.NoError(t, err)

				expected := []Bar{
					{Integer: 42, String: "Olivie Smith"},
					{Integer: 42, String: "Olivie Smith"},
				}
				require.Equal(t, expected, result.Value)
			})

			t.Run("invalid struct slice count", func(t *testing.T) {
				type InvalidStructSlice struct {
					Value []Bar `testfill:"fill:not_a_number"`
				}

				result, err := testfill.Fill(InvalidStructSlice{})

				expectedError := "testfill: failed to set field Value: invalid slice count format: fill:not_a_number"
				require.EqualError(t, err, expectedError)
				require.Equal(t, InvalidStructSlice{}, result)
			})

			t.Run("invalid int conversion in slice", func(t *testing.T) {
				type InvalidIntSlice struct {
					Value []int `testfill:"1,not_a_number,3"`
				}

				result, err := testfill.Fill(InvalidIntSlice{})

				expectedError := "testfill: failed to set field Value: unsupported slice element type int"
				require.EqualError(t, err, expectedError)
				require.Equal(t, InvalidIntSlice{}, result)
			})

			t.Run("struct slice with element fill error", func(t *testing.T) {
				type StructWithError struct {
					InvalidField int `testfill:"not_a_number"`
				}
				type SliceWithError struct {
					Value []StructWithError `testfill:"fill:2"`
				}

				result, err := testfill.Fill(SliceWithError{})

				expectedError := "testfill: failed to set field Value: failed to fill slice element 0: testfill: failed to set field InvalidField: cannot convert \"not_a_number\" to int: strconv.ParseInt: parsing \"not_a_number\": invalid syntax"
				require.EqualError(t, err, expectedError)
				require.Equal(t, SliceWithError{}, result)
			})
		})
	})

	t.Run("map", func(t *testing.T) {
		t.Run("map of string", func(t *testing.T) {
			t.Run("fills default value", func(t *testing.T) {
				type MapTest struct {
					Value map[string]string `testfill:"key1:value1,key2:value2"`
				}

				result, err := testfill.Fill(MapTest{})
				require.NoError(t, err)

				expected := map[string]string{"key1": "value1", "key2": "value2"}
				require.Equal(t, expected, result.Value)
			})

			t.Run("does not fill when value is already filled", func(t *testing.T) {
				type MapTest struct {
					Value map[string]string `testfill:"key1:value1,key2:value2"`
				}

				custom := map[string]string{"custom": "value"}
				result, err := testfill.Fill(MapTest{Value: custom})
				require.NoError(t, err)

				require.Equal(t, custom, result.Value)
			})

			t.Run("invalid map format missing colon", func(t *testing.T) {
				type InvalidMap struct {
					Value map[string]string `testfill:"key1_value1,key2:value2"`
				}

				result, err := testfill.Fill(InvalidMap{})

				expectedError := "testfill: failed to set field Value: invalid map format: key1_value1"
				require.EqualError(t, err, expectedError)
				require.Equal(t, InvalidMap{}, result)
			})

			t.Run("invalid map format too many colons", func(t *testing.T) {
				type InvalidMap struct {
					Value map[string]string `testfill:"key1:value1:extra,key2:value2"`
				}

				result, err := testfill.Fill(InvalidMap{})

				expectedError := "testfill: failed to set field Value: invalid map format: key1:value1:extra"
				require.EqualError(t, err, expectedError)
				require.Equal(t, InvalidMap{}, result)
			})

			t.Run("int key string value map", func(t *testing.T) {
				type IntStringMapTest struct {
					Value map[int]string `testfill:"1:value1,2:value2,42:answer"`
				}

				result, err := testfill.Fill(IntStringMapTest{})
				require.NoError(t, err)

				expected := map[int]string{1: "value1", 2: "value2", 42: "answer"}
				require.Equal(t, expected, result.Value)
			})

			t.Run("string int map", func(t *testing.T) {
				type StringIntMapTest struct {
					Value map[string]int `testfill:"count:10,max:100,min:1"`
				}

				result, err := testfill.Fill(StringIntMapTest{})
				require.NoError(t, err)

				expected := map[string]int{"count": 10, "max": 100, "min": 1}
				require.Equal(t, expected, result.Value)
			})

			t.Run("int float map", func(t *testing.T) {
				type IntFloatMapTest struct {
					Value map[int]float64 `testfill:"1:1.1,2:2.5,3:3.14"`
				}

				result, err := testfill.Fill(IntFloatMapTest{})
				require.NoError(t, err)

				expected := map[int]float64{1: 1.1, 2: 2.5, 3: 3.14}
				require.Equal(t, expected, result.Value)
			})

			t.Run("string bool map", func(t *testing.T) {
				type StringBoolMapTest struct {
					Value map[string]bool `testfill:"enabled:true,debug:false,verbose:true"`
				}

				result, err := testfill.Fill(StringBoolMapTest{})
				require.NoError(t, err)

				expected := map[string]bool{"enabled": true, "debug": false, "verbose": true}
				require.Equal(t, expected, result.Value)
			})

			t.Run("struct value map with fill syntax", func(t *testing.T) {
				type StructMapTest struct {
					Value map[string]Bar `testfill:"first:fill,second:fill"`
				}

				result, err := testfill.Fill(StructMapTest{})
				require.NoError(t, err)

				expected := map[string]Bar{
					"first":  {Integer: 42, String: "Olivie Smith"},
					"second": {Integer: 42, String: "Olivie Smith"},
				}
				require.Equal(t, expected, result.Value)
			})

			t.Run("unsupported struct map key type", func(t *testing.T) {
				type UnsupportedStructMap struct {
					Value map[int]Bar `testfill:"1:fill,2:fill"`
				}

				result, err := testfill.Fill(UnsupportedStructMap{})

				expectedError := "testfill: failed to set field Value: unsupported map type int -> struct"
				require.EqualError(t, err, expectedError)
				require.Equal(t, UnsupportedStructMap{}, result)
			})

			t.Run("struct map with variant names", func(t *testing.T) {
				type TestStruct struct {
					Integer int    `testfill:"42" testfill_variant1:"100"`
					String  string `testfill:"Olivie Smith" testfill_variant1:"John Doe"`
				}

				type StructMapWithVariants struct {
					Value map[string]TestStruct `testfill:"key1:variant1,key2:fill"`
				}

				result, err := testfill.Fill(StructMapWithVariants{})
				require.NoError(t, err)

				require.Len(t, result.Value, 2)

				// key1 should use variant1
				struct1, exists := result.Value["key1"]
				require.True(t, exists)
				require.Equal(t, 100, struct1.Integer)
				require.Equal(t, "John Doe", struct1.String)

				// key2 should use default (fill)
				struct2, exists := result.Value["key2"]
				require.True(t, exists)
				require.Equal(t, 42, struct2.Integer)
				require.Equal(t, "Olivie Smith", struct2.String)
			})

			t.Run("struct map with invalid format missing colon", func(t *testing.T) {
				type InvalidFormatStructMap struct {
					Value map[string]Bar `testfill:"key1_fill,key2:fill"`
				}

				result, err := testfill.Fill(InvalidFormatStructMap{})

				expectedError := "testfill: failed to set field Value: invalid map format: key1_fill"
				require.EqualError(t, err, expectedError)
				require.Equal(t, InvalidFormatStructMap{}, result)
			})

			t.Run("struct map with value fill error", func(t *testing.T) {
				type StructWithError struct {
					InvalidField float64 `testfill:"not_a_float"`
				}
				type MapWithError struct {
					Value map[string]StructWithError `testfill:"key1:fill,key2:fill"`
				}

				result, err := testfill.Fill(MapWithError{})

				expectedError := "testfill: failed to set field Value: failed to fill map value for key key1: testfill: failed to set field InvalidField: cannot convert \"not_a_float\" to float64: strconv.ParseFloat: parsing \"not_a_float\": invalid syntax"
				require.EqualError(t, err, expectedError)
				require.Equal(t, MapWithError{}, result)
			})

			t.Run("invalid key conversion in map", func(t *testing.T) {
				type InvalidKeyMap struct {
					Value map[int]string `testfill:"not_a_number:value1,2:value2"`
				}

				result, err := testfill.Fill(InvalidKeyMap{})

				expectedError := "testfill: failed to set field Value: unsupported map type int -> string"
				require.EqualError(t, err, expectedError)
				require.Equal(t, InvalidKeyMap{}, result)
			})

			t.Run("invalid value conversion in map", func(t *testing.T) {
				type InvalidValueMap struct {
					Value map[string]int `testfill:"key1:not_a_number,key2:42"`
				}

				result, err := testfill.Fill(InvalidValueMap{})

				expectedError := "testfill: failed to set field Value: unsupported map type string -> int"
				require.EqualError(t, err, expectedError)
				require.Equal(t, InvalidValueMap{}, result)
			})
		})
	})

	t.Run("struct", func(t *testing.T) {
		t.Run("nested struct with fill tag", func(t *testing.T) {
			t.Run("recursively fills nested struct fields", func(t *testing.T) {
				foo, err := testfill.Fill(Foo{})
				require.NoError(t, err)

				expected := Bar{Integer: 42, String: "Olivie Smith"}
				require.Equal(t, expected, foo.NestedStructWithFillTag)
			})

			t.Run("fills zero fields in partially filled struct", func(t *testing.T) {
				partial := Bar{Integer: 999}
				foo, err := testfill.Fill(Foo{NestedStructWithFillTag: partial})
				require.NoError(t, err)

				expected := Bar{Integer: 999, String: "Olivie Smith"}
				require.Equal(t, expected, foo.NestedStructWithFillTag)
			})

			t.Run("does not modify fully filled struct", func(t *testing.T) {
				custom := Bar{Integer: 999, String: "custom"}
				foo, err := testfill.Fill(Foo{NestedStructWithFillTag: custom})
				require.NoError(t, err)

				require.Equal(t, custom, foo.NestedStructWithFillTag)
			})
		})

		t.Run("nested struct without fill tag", func(t *testing.T) {
			t.Run("leaves field as zero value", func(t *testing.T) {
				foo, err := testfill.Fill(Foo{})
				require.NoError(t, err)

				require.Equal(t, Bar{}, foo.NestedStructWithoutTag)
			})

			t.Run("does not modify existing struct value", func(t *testing.T) {
				custom := Bar{Integer: 999, String: "custom"}
				foo, err := testfill.Fill(Foo{NestedStructWithoutTag: custom})
				require.NoError(t, err)

				require.Equal(t, custom, foo.NestedStructWithoutTag)
			})
		})

		t.Run("nested pointer with fill tag", func(t *testing.T) {
			t.Run("creates and fills pointer when nil", func(t *testing.T) {
				foo, err := testfill.Fill(Foo{})
				require.NoError(t, err)

				expected := &Bar{Integer: 42, String: "Olivie Smith"}
				require.Equal(t, expected, foo.NestedPointerWithFillTag)
			})

			t.Run("fills existing pointer struct", func(t *testing.T) {
				custom := &Bar{}
				foo, err := testfill.Fill(Foo{NestedPointerWithFillTag: custom})
				require.NoError(t, err)

				expected := &Bar{Integer: 42, String: "Olivie Smith"}
				require.Equal(t, expected, foo.NestedPointerWithFillTag)
			})

			t.Run("fills zero fields in partially filled pointer struct", func(t *testing.T) {
				custom := &Bar{Integer: 999}
				foo, err := testfill.Fill(Foo{NestedPointerWithFillTag: custom})
				require.NoError(t, err)

				expected := &Bar{Integer: 999, String: "Olivie Smith"}
				require.Equal(t, expected, foo.NestedPointerWithFillTag)
			})
		})

		t.Run("nested pointer without fill tag", func(t *testing.T) {
			t.Run("leaves field as nil", func(t *testing.T) {
				foo, err := testfill.Fill(Foo{})
				require.NoError(t, err)

				require.Nil(t, foo.NestedPointerWithoutTag)
			})

			t.Run("does not modify existing pointer value", func(t *testing.T) {
				custom := &Bar{Integer: 999, String: "custom"}
				foo, err := testfill.Fill(Foo{NestedPointerWithoutTag: custom})
				require.NoError(t, err)

				require.Equal(t, custom, foo.NestedPointerWithoutTag)
			})
		})

		t.Run("deeply nested struct with fill tag", func(t *testing.T) {
			t.Run("recursively fills all nested levels", func(t *testing.T) {
				foo, err := testfill.Fill(Foo{})
				require.NoError(t, err)

				expected := Baz{
					Name:         "Deep Nested",
					Value:        100,
					NestedBar:    Bar{Integer: 42, String: "Olivie Smith"},
					NonFilledBar: Bar{}, // This should remain empty since no fill tag
				}
				require.Equal(t, expected, foo.DeeplyNestedWithFillTag)
			})

			t.Run("fills zero fields while preserving existing values", func(t *testing.T) {
				partial := Baz{
					Name:      "Custom Name",
					NestedBar: Bar{Integer: 555},
				}
				foo, err := testfill.Fill(Foo{DeeplyNestedWithFillTag: partial})
				require.NoError(t, err)

				expected := Baz{
					Name:         "Custom Name",
					Value:        100,
					NestedBar:    Bar{Integer: 555, String: "Olivie Smith"},
					NonFilledBar: Bar{}, // This should remain empty
				}
				require.Equal(t, expected, foo.DeeplyNestedWithFillTag)
			})
		})

		t.Run("deeply nested struct without fill tag", func(t *testing.T) {
			t.Run("leaves field as zero value", func(t *testing.T) {
				foo, err := testfill.Fill(Foo{})
				require.NoError(t, err)

				require.Equal(t, Baz{}, foo.DeeplyNestedWithoutTag)
			})

			t.Run("does not modify existing struct value", func(t *testing.T) {
				custom := Baz{Name: "Custom", Value: 999}
				foo, err := testfill.Fill(Foo{DeeplyNestedWithoutTag: custom})
				require.NoError(t, err)

				require.Equal(t, custom, foo.DeeplyNestedWithoutTag)
			})
		})

		t.Run("unsupported struct type", func(t *testing.T) {
			type CustomStruct struct {
				Field string
			}
			type UnsupportedStruct struct {
				Value CustomStruct `testfill:"some_value"`
			}

			result, err := testfill.Fill(UnsupportedStruct{})

			expectedError := "testfill: failed to set field Value: unsupported struct type testfill_test.CustomStruct"
			require.EqualError(t, err, expectedError)
			require.Equal(t, UnsupportedStruct{}, result)
		})

		t.Run("nested struct with error", func(t *testing.T) {
			type NestedWithError struct {
				InvalidInt int `testfill:"not_a_number"`
			}
			type ContainerWithError struct {
				Nested NestedWithError `testfill:"fill"`
			}

			result, err := testfill.Fill(ContainerWithError{})

			expectedError := "testfill: failed to fill nested struct Nested: testfill: failed to set field InvalidInt: cannot convert \"not_a_number\" to int: strconv.ParseInt: parsing \"not_a_number\": invalid syntax"
			require.EqualError(t, err, expectedError)
			require.Equal(t, ContainerWithError{}, result)
		})

		t.Run("nested struct pointer with error", func(t *testing.T) {
			type NestedWithError struct {
				InvalidBool bool `testfill:"not_a_bool"`
			}
			type ContainerWithError struct {
				NestedPtr *NestedWithError `testfill:"fill"`
			}

			result, err := testfill.Fill(ContainerWithError{})

			expectedError := "testfill: failed to fill nested struct pointer NestedPtr: testfill: failed to set field InvalidBool: cannot convert \"not_a_bool\" to bool: strconv.ParseBool: parsing \"not_a_bool\": invalid syntax"
			require.EqualError(t, err, expectedError)
			require.Equal(t, ContainerWithError{}, result)
		})
	})

	t.Run("factory", func(t *testing.T) {
		t.Run("custom type with factory function", func(t *testing.T) {
			t.Run("fills using factory function when zero value", func(t *testing.T) {
				type CustomFactoryTest struct {
					Value CustomVO `testfill:"factory:NewCustomVO"`
				}

				result, err := testfill.Fill(CustomFactoryTest{})
				require.NoError(t, err)

				expected := CustomVO{privateField: "factory default"}
				require.Equal(t, expected, result.Value)
			})

			t.Run("does not modify existing custom value", func(t *testing.T) {
				type CustomFactoryTest struct {
					Value CustomVO `testfill:"factory:NewCustomVO"`
				}

				custom := CustomVO{privateField: "existing value"}
				result, err := testfill.Fill(CustomFactoryTest{Value: custom})
				require.NoError(t, err)

				require.Equal(t, custom, result.Value)
			})
		})

		t.Run("custom type with factory function and arguments", func(t *testing.T) {
			t.Run("fills using factory function with argument when zero value", func(t *testing.T) {
				type CustomFactoryWithArgTest struct {
					Value CustomVO `testfill:"factory:NewCustomVOWithArg:custom argument"`
				}

				result, err := testfill.Fill(CustomFactoryWithArgTest{})
				require.NoError(t, err)

				expected := CustomVO{privateField: "custom argument"}
				require.Equal(t, expected, result.Value)
			})

			t.Run("does not modify existing custom value with arg factory", func(t *testing.T) {
				type CustomFactoryWithArgTest struct {
					Value CustomVO `testfill:"factory:NewCustomVOWithArg:custom argument"`
				}

				custom := CustomVO{privateField: "existing value"}
				result, err := testfill.Fill(CustomFactoryWithArgTest{Value: custom})
				require.NoError(t, err)

				require.Equal(t, custom, result.Value)
			})
		})

		t.Run("custom type with factory function and multiple arguments", func(t *testing.T) {
			t.Run("fills using factory function with multiple arguments when zero value", func(t *testing.T) {
				type CustomFactoryMultiArgsTest struct {
					Value CustomVO `testfill:"factory:NewCustomVOMultiArgs:prefix:42:suffix"`
				}

				result, err := testfill.Fill(CustomFactoryMultiArgsTest{})
				require.NoError(t, err)

				expected := CustomVO{privateField: "prefix-42-suffix"}
				require.Equal(t, expected, result.Value)
			})

			t.Run("does not modify existing custom value with multi-arg factory", func(t *testing.T) {
				type CustomFactoryMultiArgsTest struct {
					Value CustomVO `testfill:"factory:NewCustomVOMultiArgs:prefix:42:suffix"`
				}

				custom := CustomVO{privateField: "existing value"}
				result, err := testfill.Fill(CustomFactoryMultiArgsTest{Value: custom})
				require.NoError(t, err)

				require.Equal(t, custom, result.Value)
			})
		})

		t.Run("time with factory function", func(t *testing.T) {
			t.Run("fills using ParseDate factory with string argument", func(t *testing.T) {
				type TimeFactoryTest struct {
					Value time.Time `testfill:"factory:ParseDate:2024-12-25"`
				}

				result, err := testfill.Fill(TimeFactoryTest{})
				require.NoError(t, err)

				expected := time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)
				require.Equal(t, expected, result.Value)
			})

			t.Run("does not modify existing date time value", func(t *testing.T) {
				type TimeFactoryTest struct {
					Value time.Time `testfill:"factory:ParseDate:2024-12-25"`
				}

				customTime := time.Date(2020, 6, 15, 12, 0, 0, 0, time.UTC)
				result, err := testfill.Fill(TimeFactoryTest{Value: customTime})
				require.NoError(t, err)

				require.Equal(t, customTime, result.Value)
			})
		})

		t.Run("factory function error handling", func(t *testing.T) {
			t.Run("when factory function panics returns error", func(t *testing.T) {
				type PanicTest struct {
					CustomVOWithPanic CustomVO `testfill:"factory:PanicFactory"`
				}

				result, err := testfill.Fill(PanicTest{})

				expectedError := "testfill: failed to set field CustomVOWithPanic: factory function panicked: this factory always panics"
				require.EqualError(t, err, expectedError)
				require.Equal(t, PanicTest{}, result)
			})

			t.Run("unregistered factory function", func(t *testing.T) {
				type UnregisteredFactory struct {
					Value CustomVO `testfill:"factory:NonExistentFactory"`
				}

				result, err := testfill.Fill(UnregisteredFactory{})

				expectedError := "testfill: failed to set field Value: factory function NonExistentFactory not found"
				require.EqualError(t, err, expectedError)
				require.Equal(t, UnregisteredFactory{}, result)
			})

			t.Run("wrong argument count", func(t *testing.T) {
				testfill.RegisterFactory("NoArgsFactory", func() CustomVO {
					return CustomVO{}
				})

				t.Run("too many arguments", func(t *testing.T) {
					type TooManyArgs struct {
						Value CustomVO `testfill:"factory:NoArgsFactory:extra:arg"`
					}

					result, err := testfill.Fill(TooManyArgs{})

					expectedError := "testfill: failed to set field Value: factory function NoArgsFactory expects 0 arguments, got 2"
					require.EqualError(t, err, expectedError)
					require.Equal(t, TooManyArgs{}, result)
				})

				t.Run("too few arguments", func(t *testing.T) {
					type TooFewArgs struct {
						Value CustomVO `testfill:"factory:NewCustomVOWithArg"`
					}

					result, err := testfill.Fill(TooFewArgs{})

					expectedError := "testfill: failed to set field Value: factory function NewCustomVOWithArg expects 1 arguments, got 0"
					require.EqualError(t, err, expectedError)
					require.Equal(t, TooFewArgs{}, result)
				})
			})

			t.Run("wrong return type", func(t *testing.T) {
				testfill.RegisterFactory("WrongReturnType", func() string {
					return "not a CustomVO"
				})

				type WrongReturnType struct {
					Value CustomVO `testfill:"factory:WrongReturnType"`
				}

				result, err := testfill.Fill(WrongReturnType{})

				expectedError := "testfill: failed to set field Value: factory function WrongReturnType returns string, but field expects testfill_test.CustomVO"
				require.EqualError(t, err, expectedError)
				require.Equal(t, WrongReturnType{}, result)
			})

			t.Run("multiple return values", func(t *testing.T) {
				testfill.RegisterFactory("MultipleReturns", func() (CustomVO, error) {
					return CustomVO{}, nil
				})

				type MultipleReturns struct {
					Value CustomVO `testfill:"factory:MultipleReturns"`
				}

				result, err := testfill.Fill(MultipleReturns{})

				expectedError := "testfill: failed to set field Value: factory function MultipleReturns must return exactly one value"
				require.EqualError(t, err, expectedError)
				require.Equal(t, MultipleReturns{}, result)
			})

			t.Run("argument conversion errors", func(t *testing.T) {
				testfill.RegisterFactory("IntArgFactory", func(num int) CustomVO {
					return CustomVO{}
				})

				t.Run("invalid int conversion", func(t *testing.T) {
					type InvalidIntArg struct {
						Value CustomVO `testfill:"factory:IntArgFactory:not_a_number"`
					}

					result, err := testfill.Fill(InvalidIntArg{})

					expectedError := "testfill: failed to set field Value: factory function IntArgFactory argument 0: cannot convert \"not_a_number\" to int: strconv.ParseInt: parsing \"not_a_number\": invalid syntax"
					require.EqualError(t, err, expectedError)
					require.Equal(t, InvalidIntArg{}, result)
				})

				testfill.RegisterFactory("FloatArgFactory", func(num float64) CustomVO {
					return CustomVO{}
				})

				t.Run("invalid float conversion", func(t *testing.T) {
					type InvalidFloatArg struct {
						Value CustomVO `testfill:"factory:FloatArgFactory:not_a_float"`
					}

					result, err := testfill.Fill(InvalidFloatArg{})

					expectedError := "testfill: failed to set field Value: factory function FloatArgFactory argument 0: cannot convert \"not_a_float\" to float64: strconv.ParseFloat: parsing \"not_a_float\": invalid syntax"
					require.EqualError(t, err, expectedError)
					require.Equal(t, InvalidFloatArg{}, result)
				})

				testfill.RegisterFactory("BoolArgFactory", func(flag bool) CustomVO {
					return CustomVO{}
				})

				t.Run("invalid bool conversion", func(t *testing.T) {
					type InvalidBoolArg struct {
						Value CustomVO `testfill:"factory:BoolArgFactory:not_a_bool"`
					}

					result, err := testfill.Fill(InvalidBoolArg{})

					expectedError := "testfill: failed to set field Value: factory function BoolArgFactory argument 0: cannot convert \"not_a_bool\" to bool: strconv.ParseBool: parsing \"not_a_bool\": invalid syntax"
					require.EqualError(t, err, expectedError)
					require.Equal(t, InvalidBoolArg{}, result)
				})

				testfill.RegisterFactory("UintArgFactory", func(num uint) CustomVO {
					return CustomVO{}
				})

				t.Run("valid uint conversion", func(t *testing.T) {
					type ValidUintArg struct {
						Value CustomVO `testfill:"factory:UintArgFactory:42"`
					}

					result, err := testfill.Fill(ValidUintArg{})

					require.NoError(t, err)
					require.Equal(t, CustomVO{}, result.Value)
				})

				t.Run("invalid uint conversion", func(t *testing.T) {
					type InvalidUintArg struct {
						Value CustomVO `testfill:"factory:UintArgFactory:not_a_number"`
					}

					result, err := testfill.Fill(InvalidUintArg{})

					expectedError := "testfill: failed to set field Value: factory function UintArgFactory argument 0: cannot convert \"not_a_number\" to uint: strconv.ParseUint: parsing \"not_a_number\": invalid syntax"
					require.EqualError(t, err, expectedError)
					require.Equal(t, InvalidUintArg{}, result)
				})

				t.Run("valid float conversion", func(t *testing.T) {
					type ValidFloatArg struct {
						Value CustomVO `testfill:"factory:FloatArgFactory:99.99"`
					}

					result, err := testfill.Fill(ValidFloatArg{})

					require.NoError(t, err)
					require.Equal(t, CustomVO{}, result.Value)
				})

				t.Run("valid bool conversion", func(t *testing.T) {
					type ValidBoolArg struct {
						Value CustomVO `testfill:"factory:BoolArgFactory:true"`
					}

					result, err := testfill.Fill(ValidBoolArg{})

					require.NoError(t, err)
					require.Equal(t, CustomVO{}, result.Value)
				})
			})
		})
	})

	t.Run("unexported fields", func(t *testing.T) {
		t.Run("skips unexported fields", func(t *testing.T) {
			type UnexportedFieldStruct struct {
				ExportedField   string `testfill:"exported value"`
				unexportedField string `testfill:"this should be ignored"`
			}

			result, err := testfill.Fill(UnexportedFieldStruct{})
			require.NoError(t, err)

			require.Equal(t, "exported value", result.ExportedField)
			require.Equal(t, "", result.unexportedField) // Should remain zero value
		})

		t.Run("handles mix of exported and unexported fields", func(t *testing.T) {
			type MixedFieldStruct struct {
				Field1 string `testfill:"value1"`
				field2 int    `testfill:"42"`
				Field3 bool   `testfill:"true"`
				field4 string
			}

			result, err := testfill.Fill(MixedFieldStruct{})
			require.NoError(t, err)

			require.Equal(t, "value1", result.Field1)
			require.Equal(t, 0, result.field2) // Should remain zero value
			require.Equal(t, true, result.Field3)
			require.Equal(t, "", result.field4) // Should remain zero value
		})
	})

	t.Run("edge cases", func(t *testing.T) {
		t.Run("embedded struct without fill tag", func(t *testing.T) {
			type Embedded struct {
				EmbeddedField string `testfill:"embedded value"`
			}
			type ContainerStruct struct {
				Embedded
				OtherField string `testfill:"other value"`
			}

			result, err := testfill.Fill(ContainerStruct{})
			require.NoError(t, err)

			// Embedded struct fields are not filled unless the embedded struct itself has a fill tag
			require.Equal(t, "", result.EmbeddedField)
			require.Equal(t, "other value", result.OtherField)
		})

		t.Run("embedded struct with fill tag", func(t *testing.T) {
			type Embedded struct {
				EmbeddedField string `testfill:"embedded value"`
			}
			type ContainerStruct struct {
				Embedded   `testfill:"fill"`
				OtherField string `testfill:"other value"`
			}

			result, err := testfill.Fill(ContainerStruct{})
			require.NoError(t, err)

			// When embedded struct has fill tag, its fields are filled recursively
			require.Equal(t, "embedded value", result.EmbeddedField)
			require.Equal(t, "other value", result.OtherField)
		})

		t.Run("handles anonymous fields", func(t *testing.T) {
			type AnonymousStruct struct {
				string `testfill:"anonymous string"`
			}

			result, err := testfill.Fill(AnonymousStruct{})
			require.NoError(t, err)

			// Anonymous fields cannot be set via reflection
			require.Equal(t, "", result.string)
		})

		t.Run("handles interface fields", func(t *testing.T) {
			type InterfaceStruct struct {
				// Interface fields start as nil (zero value)
				Data interface{} `testfill:"fill"`
			}

			result, err := testfill.Fill(InterfaceStruct{})
			require.NoError(t, err)

			// Interface fields without specific type cannot be filled
			require.Nil(t, result.Data)
		})
	})

	t.Run("invalid types", func(t *testing.T) {
		t.Run("passing int returns error", func(t *testing.T) {
			result, err := testfill.Fill(42)

			require.EqualError(t, err, "testfill: expected struct, got int")
			require.Equal(t, 0, result)
		})

		t.Run("passing string returns error", func(t *testing.T) {
			result, err := testfill.Fill("hello")

			require.EqualError(t, err, "testfill: expected struct, got string")
			require.Equal(t, "", result)
		})

		t.Run("passing slice returns error", func(t *testing.T) {
			result, err := testfill.Fill([]int{1, 2, 3})

			require.EqualError(t, err, "testfill: expected struct, got []int")
			require.Nil(t, result)
		})
	})

	t.Run("named variants", func(t *testing.T) {
		t.Run("basic named variants", func(t *testing.T) {
			type User struct {
				Name string `testfill:"John" testfill_admin:"Jane" testfill_guest:"Bob"`
				Age  int    `testfill:"25" testfill_admin:"30" testfill_guest:"35"`
				Role string `testfill:"user" testfill_admin:"admin" testfill_guest:"guest"`
			}

			type UserList struct {
				Users []User `testfill:"variants:default,admin,guest"`
			}

			result, err := testfill.Fill(UserList{})
			require.NoError(t, err)

			require.Len(t, result.Users, 3)

			// First user (default)
			require.Equal(t, "John", result.Users[0].Name)
			require.Equal(t, 25, result.Users[0].Age)
			require.Equal(t, "user", result.Users[0].Role)

			// Second user (admin)
			require.Equal(t, "Jane", result.Users[1].Name)
			require.Equal(t, 30, result.Users[1].Age)
			require.Equal(t, "admin", result.Users[1].Role)

			// Third user (guest)
			require.Equal(t, "Bob", result.Users[2].Name)
			require.Equal(t, 35, result.Users[2].Age)
			require.Equal(t, "guest", result.Users[2].Role)
		})

		t.Run("partial variant coverage", func(t *testing.T) {
			type User struct {
				Name string `testfill:"John" testfill_admin:"Jane"`                         // Only has admin variant
				Age  int    `testfill:"25"`                                                 // Only has default
				Role string `testfill:"user" testfill_admin:"admin" testfill_guest:"guest"` // Has all variants
			}

			type UserList struct {
				Users []User `testfill:"variants:default,admin,guest"`
			}

			result, err := testfill.Fill(UserList{})
			require.NoError(t, err)

			require.Len(t, result.Users, 3)

			// First user (default)
			require.Equal(t, "John", result.Users[0].Name)
			require.Equal(t, 25, result.Users[0].Age)
			require.Equal(t, "user", result.Users[0].Role)

			// Second user (admin)
			require.Equal(t, "Jane", result.Users[1].Name)
			require.Equal(t, 25, result.Users[1].Age) // Falls back to default
			require.Equal(t, "admin", result.Users[1].Role)

			// Third user (guest)
			require.Equal(t, "John", result.Users[2].Name) // Falls back to default
			require.Equal(t, 25, result.Users[2].Age)      // Falls back to default
			require.Equal(t, "guest", result.Users[2].Role)
		})

		t.Run("nested structs with variants", func(t *testing.T) {
			type Address struct {
				Street string `testfill:"123 Main St" testfill_work:"456 Office Blvd"`
				City   string `testfill:"New York" testfill_work:"Boston"`
			}

			type Person struct {
				Name    string  `testfill:"John" testfill_manager:"Jane"`
				Address Address `testfill:"fill"`
			}

			type PersonList struct {
				People []Person `testfill:"variants:default,manager"`
			}

			result, err := testfill.Fill(PersonList{})
			require.NoError(t, err)

			require.Len(t, result.People, 2)

			// First person (default)
			require.Equal(t, "John", result.People[0].Name)
			require.Equal(t, "123 Main St", result.People[0].Address.Street)
			require.Equal(t, "New York", result.People[0].Address.City)

			// Second person (manager)
			require.Equal(t, "Jane", result.People[1].Name)
			require.Equal(t, "123 Main St", result.People[1].Address.Street) // Nested struct uses default since no work variant
			require.Equal(t, "New York", result.People[1].Address.City)
		})

		t.Run("single variant", func(t *testing.T) {
			type User struct {
				Name string `testfill:"John" testfill_admin:"Jane"`
				Role string `testfill:"user" testfill_admin:"admin"`
			}

			type UserList struct {
				Users []User `testfill:"variants:admin"`
			}

			result, err := testfill.Fill(UserList{})
			require.NoError(t, err)

			require.Len(t, result.Users, 1)
			require.Equal(t, "Jane", result.Users[0].Name)
			require.Equal(t, "admin", result.Users[0].Role)
		})

		t.Run("whitespace handling in variants", func(t *testing.T) {
			type User struct {
				Name string `testfill:"John" testfill_admin:"Jane"`
			}

			type UserList struct {
				Users []User `testfill:"variants: default , admin , guest "`
			}

			result, err := testfill.Fill(UserList{})
			require.NoError(t, err)

			require.Len(t, result.Users, 3)
			require.Equal(t, "John", result.Users[0].Name)
			require.Equal(t, "Jane", result.Users[1].Name)
			require.Equal(t, "John", result.Users[2].Name) // guest falls back to default
		})

		t.Run("empty variant list", func(t *testing.T) {
			type User struct {
				Name string `testfill:"John"`
			}

			type UserList struct {
				Users []User `testfill:"variants:"`
			}

			result, err := testfill.Fill(UserList{})
			require.NoError(t, err)

			require.Len(t, result.Users, 1)
			require.Equal(t, "John", result.Users[0].Name)
		})

		t.Run("preserves existing non-zero values", func(t *testing.T) {
			type User struct {
				Name string `testfill:"John" testfill_admin:"Jane"`
				Age  int    `testfill:"25" testfill_admin:"30"`
			}

			type UserList struct {
				Users []User `testfill:"variants:default,admin"`
			}

			// Pre-fill first user partially
			input := UserList{
				Users: []User{
					{Name: "CustomName"}, // Age should still get filled
				},
			}

			result, err := testfill.Fill(input)
			require.NoError(t, err)

			// Should preserve existing slice, not create new one
			require.Len(t, result.Users, 1)
			require.Equal(t, "CustomName", result.Users[0].Name) // Preserved
			require.Equal(t, 0, result.Users[0].Age)             // Not filled since slice was pre-existing
		})

		t.Run("map with custom key=variant pairs", func(t *testing.T) {
			type User struct {
				Name string `testfill:"John" testfill_admin:"Jane" testfill_guest:"Bob"`
				Age  int    `testfill:"25" testfill_admin:"30" testfill_guest:"35"`
				Role string `testfill:"user" testfill_admin:"admin" testfill_guest:"guest"`
			}

			type UserMap struct {
				Users map[string]User `testfill:"variants:regular_user=default,system_admin=admin,site_visitor=guest"`
			}

			result, err := testfill.Fill(UserMap{})
			require.NoError(t, err)

			require.Len(t, result.Users, 3)

			// regular_user should have default variant
			regularUser, exists := result.Users["regular_user"]
			require.True(t, exists)
			require.Equal(t, "John", regularUser.Name)
			require.Equal(t, 25, regularUser.Age)
			require.Equal(t, "user", regularUser.Role)

			// system_admin should have admin variant
			systemAdmin, exists := result.Users["system_admin"]
			require.True(t, exists)
			require.Equal(t, "Jane", systemAdmin.Name)
			require.Equal(t, 30, systemAdmin.Age)
			require.Equal(t, "admin", systemAdmin.Role)

			// site_visitor should have guest variant
			siteVisitor, exists := result.Users["site_visitor"]
			require.True(t, exists)
			require.Equal(t, "Bob", siteVisitor.Name)
			require.Equal(t, 35, siteVisitor.Age)
			require.Equal(t, "guest", siteVisitor.Role)
		})

		t.Run("map with mixed custom keys and whitespace", func(t *testing.T) {
			type User struct {
				Name string `testfill:"John" testfill_admin:"Jane"`
				Role string `testfill:"user" testfill_admin:"admin"`
			}

			type UserMap struct {
				Users map[string]User `testfill:"variants: primary_user = default , admin_user = admin "`
			}

			result, err := testfill.Fill(UserMap{})
			require.NoError(t, err)

			require.Len(t, result.Users, 2)

			primaryUser, exists := result.Users["primary_user"]
			require.True(t, exists)
			require.Equal(t, "John", primaryUser.Name)
			require.Equal(t, "user", primaryUser.Role)

			adminUser, exists := result.Users["admin_user"]
			require.True(t, exists)
			require.Equal(t, "Jane", adminUser.Name)
			require.Equal(t, "admin", adminUser.Role)
		})

		t.Run("map with invalid key=variant format", func(t *testing.T) {
			type User struct {
				Name string `testfill:"John"`
			}

			type UserMap struct {
				Users map[string]User `testfill:"variants:key1=admin,invalid_format,key3=guest"`
			}

			result, err := testfill.Fill(UserMap{})

			expectedError := "testfill: failed to set field Users: invalid key=variant format: invalid_format (expected format: key=variant)"
			require.EqualError(t, err, expectedError)
			require.Equal(t, UserMap{}, result)
		})

		t.Run("map with custom keys and partial variant coverage", func(t *testing.T) {
			type User struct {
				Name string `testfill:"John" testfill_admin:"Jane"`                         // Only has admin variant
				Age  int    `testfill:"25"`                                                 // Only has default
				Role string `testfill:"user" testfill_admin:"admin" testfill_guest:"guest"` // Has all variants
			}

			type UserMap struct {
				Users map[string]User `testfill:"variants:main_user=default,power_user=admin,visitor=guest"`
			}

			result, err := testfill.Fill(UserMap{})
			require.NoError(t, err)

			require.Len(t, result.Users, 3)

			// main_user with default variant
			mainUser, exists := result.Users["main_user"]
			require.True(t, exists)
			require.Equal(t, "John", mainUser.Name)
			require.Equal(t, 25, mainUser.Age)
			require.Equal(t, "user", mainUser.Role)

			// power_user with admin variant
			powerUser, exists := result.Users["power_user"]
			require.True(t, exists)
			require.Equal(t, "Jane", powerUser.Name)
			require.Equal(t, 25, powerUser.Age) // Falls back to default
			require.Equal(t, "admin", powerUser.Role)

			// visitor with guest variant
			visitor, exists := result.Users["visitor"]
			require.True(t, exists)
			require.Equal(t, "John", visitor.Name) // Falls back to default
			require.Equal(t, 25, visitor.Age)      // Falls back to default
			require.Equal(t, "guest", visitor.Role)
		})

		t.Run("map with specific key-variant pairs", func(t *testing.T) {
			type User struct {
				Name string `testfill:"John" testfill_admin:"Jane" testfill_guest:"Bob"`
				Role string `testfill:"user" testfill_admin:"admin" testfill_guest:"guest"`
			}

			type UserMap struct {
				Users map[string]User `testfill:"alice:admin,bob:guest,charlie:default"`
			}

			result, err := testfill.Fill(UserMap{})
			require.NoError(t, err)

			require.Len(t, result.Users, 3)

			// alice should have admin variant
			alice, exists := result.Users["alice"]
			require.True(t, exists)
			require.Equal(t, "Jane", alice.Name)
			require.Equal(t, "admin", alice.Role)

			// bob should have guest variant
			bob, exists := result.Users["bob"]
			require.True(t, exists)
			require.Equal(t, "Bob", bob.Name)
			require.Equal(t, "guest", bob.Role)

			// charlie should have default variant
			charlie, exists := result.Users["charlie"]
			require.True(t, exists)
			require.Equal(t, "John", charlie.Name)
			require.Equal(t, "user", charlie.Role)
		})

		t.Run("map compatibility with fill syntax", func(t *testing.T) {
			type User struct {
				Name string `testfill:"John"`
				Age  int    `testfill:"25"`
			}

			type UserMap struct {
				Users map[string]User `testfill:"user1:fill,user2:fill"`
			}

			result, err := testfill.Fill(UserMap{})
			require.NoError(t, err)

			require.Len(t, result.Users, 2)

			user1, exists := result.Users["user1"]
			require.True(t, exists)
			require.Equal(t, "John", user1.Name)
			require.Equal(t, 25, user1.Age)

			user2, exists := result.Users["user2"]
			require.True(t, exists)
			require.Equal(t, "John", user2.Name)
			require.Equal(t, 25, user2.Age)
		})

		t.Run("map with partial variant coverage", func(t *testing.T) {
			type User struct {
				Name string `testfill:"John" testfill_admin:"Jane"`                         // Only has admin variant
				Age  int    `testfill:"25"`                                                 // Only has default
				Role string `testfill:"user" testfill_admin:"admin" testfill_guest:"guest"` // Has all variants
			}

			type UserMap struct {
				Users map[string]User `testfill:"user1:admin,user2:guest"`
			}

			result, err := testfill.Fill(UserMap{})
			require.NoError(t, err)

			require.Len(t, result.Users, 2)

			// user1 with admin variant
			user1, exists := result.Users["user1"]
			require.True(t, exists)
			require.Equal(t, "Jane", user1.Name)
			require.Equal(t, 25, user1.Age) // Falls back to default
			require.Equal(t, "admin", user1.Role)

			// user2 with guest variant
			user2, exists := result.Users["user2"]
			require.True(t, exists)
			require.Equal(t, "John", user2.Name) // Falls back to default
			require.Equal(t, 25, user2.Age)      // Falls back to default
			require.Equal(t, "guest", user2.Role)
		})
	})

	t.Run("json unmarshal", func(t *testing.T) {
		t.Run("various types", func(t *testing.T) {
			type TestJSON struct {
				String    string         `testfill:"unmarshal:\"hello world\""`
				Int       int            `testfill:"unmarshal:42"`
				Float     float64        `testfill:"unmarshal:99.99"`
				Bool      bool           `testfill:"unmarshal:true"`
				StringPtr *string        `testfill:"unmarshal:\"hello\""`
				NullPtr   *string        `testfill:"unmarshal:null"`
				Slice     []string       `testfill:"unmarshal:[\"a\",\"b\",\"c\"]"`
				Map       map[string]int `testfill:"unmarshal:{\"x\":1,\"y\":2}"`
				Interface interface{}    `testfill:"unmarshal:{\"key\":\"value\"}"`
				Time      time.Time      `testfill:"unmarshal:\"2024-01-15T10:30:00Z\""`
			}

			result, err := testfill.Fill(TestJSON{})
			require.NoError(t, err)

			// Verify all fields
			require.Equal(t, "hello world", result.String)
			require.Equal(t, 42, result.Int)
			require.Equal(t, 99.99, result.Float)
			require.Equal(t, true, result.Bool)
			require.NotNil(t, result.StringPtr)
			require.Equal(t, "hello", *result.StringPtr)
			require.Nil(t, result.NullPtr)
			require.Equal(t, []string{"a", "b", "c"}, result.Slice)
			require.Equal(t, map[string]int{"x": 1, "y": 2}, result.Map)

			m, ok := result.Interface.(map[string]interface{})
			require.True(t, ok)
			require.Equal(t, "value", m["key"])

			expected, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
			require.Equal(t, expected, result.Time)
		})

		t.Run("complex struct", func(t *testing.T) {
			type Address struct {
				Street string `json:"street"`
				City   string `json:"city"`
			}
			type Person struct {
				Name    string   `json:"name"`
				Age     int      `json:"age"`
				Address Address  `json:"address"`
				Tags    []string `json:"tags"`
			}
			type TestStruct struct {
				Person Person `testfill:"unmarshal:{\"name\":\"Alice\",\"age\":30,\"address\":{\"street\":\"123 Main\",\"city\":\"NYC\"},\"tags\":[\"dev\",\"lead\"]}"`
			}

			result, err := testfill.Fill(TestStruct{})
			require.NoError(t, err)

			require.Equal(t, "Alice", result.Person.Name)
			require.Equal(t, 30, result.Person.Age)
			require.Equal(t, "123 Main", result.Person.Address.Street)
			require.Equal(t, "NYC", result.Person.Address.City)
			require.Equal(t, []string{"dev", "lead"}, result.Person.Tags)
		})

		t.Run("preserves existing values", func(t *testing.T) {
			type TestPreserve struct {
				Value string  `testfill:"unmarshal:\"new\""`
				Ptr   *string `testfill:"unmarshal:\"new\""`
			}

			existing := "existing"
			input := TestPreserve{
				Value: "existing",
				Ptr:   &existing,
			}

			result, err := testfill.Fill(input)
			require.NoError(t, err)
			require.Equal(t, "existing", result.Value)
			require.Equal(t, "existing", *result.Ptr)
		})

		t.Run("error cases", func(t *testing.T) {
			tests := []struct {
				name     string
				input    interface{}
				errorMsg string
			}{
				{
					name: "invalid JSON",
					input: struct {
						Value map[string]string `testfill:"unmarshal:{invalid}"`
					}{},
					errorMsg: "failed to unmarshal JSON",
				},
				{
					name: "type mismatch",
					input: struct {
						Value int `testfill:"unmarshal:\"not a number\""`
					}{},
					errorMsg: "failed to unmarshal JSON",
				},
				{
					name: "invalid array element",
					input: struct {
						Value []int `testfill:"unmarshal:[1,\"two\",3]"`
					}{},
					errorMsg: "failed to unmarshal JSON",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					_, err := testfill.Fill(tt.input)
					require.Error(t, err)
					require.Contains(t, err.Error(), tt.errorMsg)
				})
			}
		})
	})
}
