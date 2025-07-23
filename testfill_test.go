package testfill_test

import (
	"testing"
	"time"

	"github.com/fabioelizandro/testfill"
	"github.com/stretchr/testify/require"
)

type Foo struct {
	Integer                    int               `testfill:"42"`
	String                     string            `testfill:"John Doe"`
	Boolean                    bool              `testfill:"true"`
	Float                      float64           `testfill:"99.99"`
	StdVO                      time.Time         `testfill:"2023-01-15T10:30:00Z"`
	ArrayOfString              []string          `testfill:"tag1,tag2,tag3"`
	MapOfString                map[string]string `testfill:"key1:value1,key2:value2"`
	NestedStructWithFillTag    Bar               `testfill:"fill"`
	NestedStructWithoutTag     Bar
	NestedPointerWithFillTag   *Bar              `testfill:"fill"`
	NestedPointerWithoutTag    *Bar
	DeeplyNestedWithFillTag    Baz               `testfill:"fill"`
	DeeplyNestedWithoutTag     Baz
}

type Bar struct {
	Integer int    `testfill:"42"`
	String  string `testfill:"Olivie Smith"`
}

type Baz struct {
	Name        string `testfill:"Deep Nested"`
	Value       int    `testfill:"100"`
	NestedBar   Bar    `testfill:"fill"`
	NonFilledBar Bar
}

func TestTestfill(t *testing.T) {
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
	})

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
	})

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
	})

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
				Name:        "Deep Nested",
				Value:       100,
				NestedBar:   Bar{Integer: 42, String: "Olivie Smith"},
				NonFilledBar: Bar{}, // This should remain empty since no fill tag
			}
			require.Equal(t, expected, foo.DeeplyNestedWithFillTag)
		})

		t.Run("fills zero fields while preserving existing values", func(t *testing.T) {
			partial := Baz{
				Name:        "Custom Name",
				NestedBar:   Bar{Integer: 555},
			}
			foo, err := testfill.Fill(Foo{DeeplyNestedWithFillTag: partial})
			require.NoError(t, err)

			expected := Baz{
				Name:        "Custom Name",
				Value:       100,
				NestedBar:   Bar{Integer: 555, String: "Olivie Smith"},
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
}
