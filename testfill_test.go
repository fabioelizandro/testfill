package testfill_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/fabioelizandro/testfill"
	"github.com/stretchr/testify/require"
)

type Foo struct {
	Integer                  int               `testfill:"42"`
	String                   string            `testfill:"John Doe"`
	Boolean                  bool              `testfill:"true"`
	Float                    float64           `testfill:"99.99"`
	StdVO                    time.Time         `testfill:"2023-01-15T10:30:00Z"`
	TimeWithRFCFactory       time.Time         `testfill:"factory:ParseDate:2024-12-25"`
	ArrayOfString            []string          `testfill:"tag1,tag2,tag3"`
	MapOfString              map[string]string `testfill:"key1:value1,key2:value2"`
	NestedStructWithFillTag  Bar               `testfill:"fill"`
	NestedStructWithoutTag   Bar
	NestedPointerWithFillTag *Bar `testfill:"fill"`
	NestedPointerWithoutTag  *Bar
	DeeplyNestedWithFillTag  Baz `testfill:"fill"`
	DeeplyNestedWithoutTag   Baz
	CustomVO                 CustomVO `testfill:"factory:NewCustomVO"`
	CustomVOWithArg          CustomVO `testfill:"factory:NewCustomVOWithArg:custom argument"`
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
		t.Run("fills default value", func(t *testing.T) {
			foo, err := testfill.Fill(Foo{})
			require.NoError(t, err)

			require.Equal(t, 42, foo.Integer)
		})

		t.Run("does not fill when value is already filled", func(t *testing.T) {
			foo, err := testfill.Fill(Foo{Integer: 21})
			require.NoError(t, err)

			require.Equal(t, 21, foo.Integer)
		})

		t.Run("invalid int tag", func(t *testing.T) {
			type InvalidInt struct {
				Value int `testfill:"not_a_number"`
			}

			result, err := testfill.Fill(InvalidInt{})

			expectedError := "testfill: failed to set field Value: strconv.ParseInt: parsing \"not_a_number\": invalid syntax"
			require.EqualError(t, err, expectedError)
			require.Equal(t, InvalidInt{}, result)
		})
	})

	t.Run("string", func(t *testing.T) {
		t.Run("fills default value", func(t *testing.T) {
			foo, err := testfill.Fill(Foo{})
			require.NoError(t, err)

			require.Equal(t, "John Doe", foo.String)
		})

		t.Run("does not fill when value is already filled", func(t *testing.T) {
			foo, err := testfill.Fill(Foo{String: "Mary"})
			require.NoError(t, err)

			require.Equal(t, "Mary", foo.String)
		})
	})

	t.Run("boolean", func(t *testing.T) {
		t.Run("fills default value", func(t *testing.T) {
			foo, err := testfill.Fill(Foo{})
			require.NoError(t, err)

			require.Equal(t, true, foo.Boolean)
		})

		t.Run("does not fill when value is already filled", func(t *testing.T) {
			foo, err := testfill.Fill(Foo{Boolean: true})
			require.NoError(t, err)

			require.Equal(t, true, foo.Boolean)
		})

		t.Run("invalid bool tag", func(t *testing.T) {
			type InvalidBool struct {
				Value bool `testfill:"not_a_bool"`
			}

			result, err := testfill.Fill(InvalidBool{})

			expectedError := "testfill: failed to set field Value: strconv.ParseBool: parsing \"not_a_bool\": invalid syntax"
			require.EqualError(t, err, expectedError)
			require.Equal(t, InvalidBool{}, result)
		})
	})

	t.Run("float", func(t *testing.T) {
		t.Run("fills default value", func(t *testing.T) {
			foo, err := testfill.Fill(Foo{})
			require.NoError(t, err)

			require.Equal(t, 99.99, foo.Float)
		})

		t.Run("does not fill when value is already filled", func(t *testing.T) {
			foo, err := testfill.Fill(Foo{Float: 123.45})
			require.NoError(t, err)

			require.Equal(t, 123.45, foo.Float)
		})

		t.Run("invalid float tag", func(t *testing.T) {
			type InvalidFloat struct {
				Value float64 `testfill:"not_a_float"`
			}

			result, err := testfill.Fill(InvalidFloat{})

			expectedError := "testfill: failed to set field Value: strconv.ParseFloat: parsing \"not_a_float\": invalid syntax"
			require.EqualError(t, err, expectedError)
			require.Equal(t, InvalidFloat{}, result)
		})
	})

	t.Run("time", func(t *testing.T) {
		t.Run("fills default value", func(t *testing.T) {
			foo, err := testfill.Fill(Foo{})
			require.NoError(t, err)

			expected, _ := time.Parse(time.RFC3339, "2023-01-15T10:30:00Z")
			require.Equal(t, expected, foo.StdVO)
		})

		t.Run("does not fill when value is already filled", func(t *testing.T) {
			customTime, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
			foo, err := testfill.Fill(Foo{StdVO: customTime})
			require.NoError(t, err)

			require.Equal(t, customTime, foo.StdVO)
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
				foo, err := testfill.Fill(Foo{})
				require.NoError(t, err)

				require.Equal(t, []string{"tag1", "tag2", "tag3"}, foo.ArrayOfString)
			})

			t.Run("does not fill when value is already filled", func(t *testing.T) {
				custom := []string{"custom1", "custom2"}
				foo, err := testfill.Fill(Foo{ArrayOfString: custom})
				require.NoError(t, err)

				require.Equal(t, custom, foo.ArrayOfString)
			})

			t.Run("unsupported slice type", func(t *testing.T) {
				type UnsupportedSlice struct {
					Value []int `testfill:"1,2,3"`
				}

				result, err := testfill.Fill(UnsupportedSlice{})

				expectedError := "testfill: failed to set field Value: only string slices are supported"
				require.EqualError(t, err, expectedError)
				require.Equal(t, UnsupportedSlice{}, result)
			})
		})
	})

	t.Run("map", func(t *testing.T) {
		t.Run("map of string", func(t *testing.T) {
			t.Run("fills default value", func(t *testing.T) {
				foo, err := testfill.Fill(Foo{})
				require.NoError(t, err)

				expected := map[string]string{"key1": "value1", "key2": "value2"}
				require.Equal(t, expected, foo.MapOfString)
			})

			t.Run("does not fill when value is already filled", func(t *testing.T) {
				custom := map[string]string{"custom": "value"}
				foo, err := testfill.Fill(Foo{MapOfString: custom})
				require.NoError(t, err)

				require.Equal(t, custom, foo.MapOfString)
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

			t.Run("unsupported map key type", func(t *testing.T) {
				type UnsupportedMap struct {
					Value map[int]string `testfill:"1:value1,2:value2"`
				}

				result, err := testfill.Fill(UnsupportedMap{})

				expectedError := "testfill: failed to set field Value: only string->string maps are supported"
				require.EqualError(t, err, expectedError)
				require.Equal(t, UnsupportedMap{}, result)
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
	})

	t.Run("factory", func(t *testing.T) {
		t.Run("custom type with factory function", func(t *testing.T) {
			t.Run("fills using factory function when zero value", func(t *testing.T) {
				foo, err := testfill.Fill(Foo{})
				require.NoError(t, err)

				expected := CustomVO{privateField: "factory default"}
				require.Equal(t, expected, foo.CustomVO)
			})

			t.Run("does not modify existing custom value", func(t *testing.T) {
				custom := CustomVO{privateField: "existing value"}
				foo, err := testfill.Fill(Foo{CustomVO: custom})
				require.NoError(t, err)

				require.Equal(t, custom, foo.CustomVO)
			})
		})

		t.Run("custom type with factory function and arguments", func(t *testing.T) {
			t.Run("fills using factory function with argument when zero value", func(t *testing.T) {
				foo, err := testfill.Fill(Foo{})
				require.NoError(t, err)

				expected := CustomVO{privateField: "custom argument"}
				require.Equal(t, expected, foo.CustomVOWithArg)
			})

			t.Run("does not modify existing custom value with arg factory", func(t *testing.T) {
				custom := CustomVO{privateField: "existing value"}
				foo, err := testfill.Fill(Foo{CustomVOWithArg: custom})
				require.NoError(t, err)

				require.Equal(t, custom, foo.CustomVOWithArg)
			})
		})

		t.Run("custom type with factory function and multiple arguments", func(t *testing.T) {
			t.Run("fills using factory function with multiple arguments when zero value", func(t *testing.T) {
				foo, err := testfill.Fill(Foo{})
				require.NoError(t, err)

				expected := CustomVO{privateField: "prefix-42-suffix"}
				require.Equal(t, expected, foo.CustomVOMultiArgs)
			})

			t.Run("does not modify existing custom value with multi-arg factory", func(t *testing.T) {
				custom := CustomVO{privateField: "existing value"}
				foo, err := testfill.Fill(Foo{CustomVOMultiArgs: custom})
				require.NoError(t, err)

				require.Equal(t, custom, foo.CustomVOMultiArgs)
			})
		})

		t.Run("time with factory function", func(t *testing.T) {
			t.Run("fills using ParseDate factory with string argument", func(t *testing.T) {
				foo, err := testfill.Fill(Foo{})
				require.NoError(t, err)

				expected := time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)
				require.Equal(t, expected, foo.TimeWithRFCFactory)
			})

			t.Run("does not modify existing date time value", func(t *testing.T) {
				customTime := time.Date(2020, 6, 15, 12, 0, 0, 0, time.UTC)
				foo, err := testfill.Fill(Foo{TimeWithRFCFactory: customTime})
				require.NoError(t, err)

				require.Equal(t, customTime, foo.TimeWithRFCFactory)
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
			})
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
}
