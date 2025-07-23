# testfill

A Go library for automatically filling struct fields with test data based on struct tags. Perfect for reducing boilerplate in tests by providing sensible defaults for struct fields.

## Features

- 🏷️ **Tag-based field filling** - Use struct tags to define default values
- 🔧 **Zero-value preservation** - Only fills fields that are zero-valued
- 🏭 **Factory functions** - Register custom functions for complex type initialization
- 🪆 **Nested struct support** - Recursively fill nested structs with the `fill` tag
- 🎯 **Type safety** - Full type checking at compile time with generics
- ⚡ **Simple API** - Just one function call to fill your structs

## Installation

```bash
go get github.com/fabioelizandro/testfill
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/fabioelizandro/testfill"
)

type User struct {
    Name  string `testfill:"John Doe"`
    Age   int    `testfill:"30"`
    Email string `testfill:"john@example.com"`
}

func main() {
    user, err := testfill.Fill(User{})
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("%+v\n", user)
    // Output: {Name:John Doe Age:30 Email:john@example.com}
}
```

## Supported Types

### Basic Types
- **Integers**: `int`, `int8`, `int16`, `int32`, `int64`
- **Unsigned integers**: `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- **Floating point**: `float32`, `float64`
- **Strings**: `string`
- **Booleans**: `bool`
- **Time**: `time.Time` (RFC3339 format)

### Complex Types
- **Slices**: All primitive types + structs (comma-separated values or `fill:count`)
- **Maps**: All combinations of primitive key/value types + struct values
- **Pointers**: Pointers to any supported type
- **Structs**: Nested structs with `fill` tag

## Usage Examples

### Basic Field Filling

```go
type Product struct {
    Name     string  `testfill:"Widget"`
    Price    float64 `testfill:"99.99"`
    InStock  bool    `testfill:"true"`
    Quantity int     `testfill:"100"`
}

product, _ := testfill.Fill(Product{})
// Result: {Name:Widget Price:99.99 InStock:true Quantity:100}
```

### Preserving Existing Values

Fields with non-zero values are not modified:

```go
type Config struct {
    Host string `testfill:"localhost"`
    Port int    `testfill:"8080"`
}

config, _ := testfill.Fill(Config{Port: 3000})
// Result: {Host:localhost Port:3000}
```

### Nested Structs

Use the `fill` tag to recursively fill nested structs:

```go
type Address struct {
    Street string `testfill:"123 Main St"`
    City   string `testfill:"New York"`
}

type Person struct {
    Name    string  `testfill:"Jane Doe"`
    Address Address `testfill:"fill"`
}

person, _ := testfill.Fill(Person{})
// Result: {Name:Jane Doe Address:{Street:123 Main St City:New York}}
```

### Slices and Maps

#### Primitive Slices
```go
type Collections struct {
    Strings []string  `testfill:"go,testing,automation"`
    Numbers []int     `testfill:"1,2,3,42"`
    Floats  []float64 `testfill:"1.1,2.5,3.14"`
    Flags   []bool    `testfill:"true,false,true"`
}

collections, _ := testfill.Fill(Collections{})
// Result: {Strings:[go testing automation] Numbers:[1 2 3 42] Floats:[1.1 2.5 3.14] Flags:[true false true]}
```

#### Struct Slices
```go
type User struct {
    Name string `testfill:"User"`
    Age  int    `testfill:"25"`
}

type UserList struct {
    Users []User `testfill:"fill:3"`
}

userList, _ := testfill.Fill(UserList{})
// Result: {Users:[{Name:User Age:25} {Name:User Age:25} {Name:User Age:25}]}
```

#### Primitive Maps
```go
type Settings struct {
    StringMap map[string]string `testfill:"version:1.0,author:john"`
    IntMap    map[string]int    `testfill:"count:10,max:100"`
    FloatMap  map[int]float64   `testfill:"1:1.1,2:2.5"`
    BoolMap   map[string]bool   `testfill:"enabled:true,debug:false"`
}

settings, _ := testfill.Fill(Settings{})
// Result: Multiple map types filled according to their tag values
```

#### Struct Value Maps
```go
type UserMap struct {
    Users map[string]User `testfill:"admin:fill,guest:fill"`
}

userMap, _ := testfill.Fill(UserMap{})
// Result: {Users:map[admin:{Name:User Age:25} guest:{Name:User Age:25}]}
```

### Time Values

```go
type Event struct {
    Name      string    `testfill:"Conference"`
    StartTime time.Time `testfill:"2024-01-15T10:30:00Z"`
}

event, _ := testfill.Fill(Event{})
// Result: {Name:Conference StartTime:2024-01-15 10:30:00 +0000 UTC}
```

## Factory Functions

Factory functions allow you to generate dynamic or complex values:

```go
// Register a factory function
testfill.RegisterFactory("uuid", func() string {
    return uuid.New().String()
})

testfill.RegisterFactory("timestamp", func() time.Time {
    return time.Now()
})

// Use in struct tags
type Document struct {
    ID        string    `testfill:"factory:uuid"`
    CreatedAt time.Time `testfill:"factory:timestamp"`
}

doc, _ := testfill.Fill(Document{})
// Result: {ID:550e8400-e29b-41d4-a716-446655440000 CreatedAt:2024-01-15 10:30:00}
```

### Factory Functions with Arguments

Factory functions can accept string arguments that are automatically converted:

```go
// Register a factory with arguments
testfill.RegisterFactory("randomInt", func(min, max int) int {
    return rand.Intn(max-min+1) + min
})

testfill.RegisterFactory("prefix", func(prefix string, length int) string {
    return fmt.Sprintf("%s-%d", prefix, length)
})

// Use with arguments (separated by colons)
type Game struct {
    Score  int    `testfill:"factory:randomInt:1:100"`
    Code   string `testfill:"factory:prefix:GAME:12345"`
}
```

## Advanced Usage

### Pointers

```go
type Settings struct {
    Debug   *bool   `testfill:"true"`
    Timeout *int    `testfill:"30"`
    Name    *string `testfill:"default"`
}

settings, _ := testfill.Fill(Settings{})
// All pointer fields will be allocated and filled
```

### Deeply Nested Structures

```go
type Database struct {
    Host string `testfill:"localhost"`
    Port int    `testfill:"5432"`
}

type Cache struct {
    Enabled bool `testfill:"true"`
    TTL     int  `testfill:"3600"`
}

type Config struct {
    Database Database `testfill:"fill"`
    Cache    Cache    `testfill:"fill"`
}

type App struct {
    Name   string `testfill:"MyApp"`
    Config Config `testfill:"fill"`
}

app, _ := testfill.Fill(App{})
// Recursively fills all nested structs marked with "fill"
```

## Error Handling

The library returns descriptive errors for common issues:

```go
type Invalid struct {
    Count int `testfill:"not_a_number"`
}

_, err := testfill.Fill(Invalid{})
// Error: testfill: failed to set field Count: cannot convert "not_a_number" to int: strconv.ParseInt: parsing "not_a_number": invalid syntax
```

Common error scenarios:
- Invalid type conversions (e.g., "abc" for an int field)
- Unsupported field types
- Factory function errors (not found, wrong arguments, wrong return type)
- Invalid struct input (passing non-struct types)

## Best Practices

1. **Use meaningful defaults** - Choose test values that make sense for your domain
2. **Register factories early** - Set up factory functions in `init()` or test setup
3. **Keep factories simple** - Factory functions should be deterministic when possible
4. **Document your tags** - Comment complex tag values for clarity

## Limitations

- Struct value maps only support string keys (use `key:fill` syntax)
- Struct slices require `fill:count` syntax (e.g., `fill:3` for 3 instances)
- Interface fields cannot be filled
- Channels and functions are not supported
- Unexported fields are ignored

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
