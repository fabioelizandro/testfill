# testfill

A Go library for automatically filling struct fields with test data based on struct tags. Perfect for reducing boilerplate in tests by providing sensible defaults for struct fields.

## Features

- 🏷️ **Tag-based field filling** - Use struct tags to define default values
- 🔧 **Zero-value preservation** - Only fills fields that are zero-valued
- 🏭 **Factory functions** - Register custom functions for complex type initialization
- 🪆 **Nested struct support** - Recursively fill nested structs with the `fill` tag
- 📋 **JSON unmarshaling** - Populate fields from JSON data with the `unmarshal:` prefix
- 🎭 **Named variants** - Create slices with different field values using variant tags
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

#### Named Variants for Different Test Data
Create slices where each item has different field values using named variants:

```go
type User struct {
    Name string `testfill:"John" testfill_admin:"Jane" testfill_guest:"Bob"`
    Age  int    `testfill:"25" testfill_admin:"30" testfill_guest:"35"`
    Role string `testfill:"user" testfill_admin:"admin" testfill_guest:"guest"`
}

type UserList struct {
    Users []User `testfill:"variants:default,admin,guest"`
}

userList, _ := testfill.Fill(UserList{})
// Result: {Users:[
//   {Name:John Age:25 Role:user},     // default variant
//   {Name:Jane Age:30 Role:admin},    // admin variant  
//   {Name:Bob Age:35 Role:guest}      // guest variant
// ]}
```

**Variant Features:**
- Use `testfill_<variant_name>` tags to define alternative values
- If a variant doesn't exist for a field, falls back to the default `testfill` tag
- Supports nested structs - variant selection propagates to nested fields
- Handles partial variant coverage gracefully
- **Slices**: Use `variants:name1,name2,name3` for auto-indexed items
- **Maps**: Use `variants:key1=variant1,key2=variant2` for custom keys with explicit naming

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

#### Struct Value Maps with Variants
Maps can also use named variants to create different struct values:

```go
type User struct {
    Name string `testfill:"John" testfill_admin:"Jane" testfill_guest:"Bob"`
    Role string `testfill:"user" testfill_admin:"admin" testfill_guest:"guest"`
}

// Option 1: Custom keys with key=variant syntax
type UserMap1 struct {
    Users map[string]User `testfill:"variants:regular_user=default,system_admin=admin,site_visitor=guest"`
}

userMap1, _ := testfill.Fill(UserMap1{})
// Result: {Users:map[regular_user:{Name:John Role:user} system_admin:{Name:Jane Role:admin} site_visitor:{Name:Bob Role:guest}]}

// Option 2: Specific keys with direct variant assignment
type UserMap2 struct {
    Users map[string]User `testfill:"alice:admin,bob:guest,charlie:default"`
}

userMap2, _ := testfill.Fill(UserMap2{})
// Result: {Users:map[alice:{Name:Jane Role:admin} bob:{Name:Bob Role:guest} charlie:{Name:John Role:user}]}
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

### JSON Unmarshaling

Use the `unmarshal:` prefix to populate fields from JSON data:

#### Basic Types
```go
type Config struct {
    Settings map[string]interface{} `testfill:"unmarshal:{\"theme\":\"dark\",\"fontSize\":14}"`
    Tags     []string              `testfill:"unmarshal:[\"go\",\"testing\",\"automation\"]"`
}

config, _ := testfill.Fill(Config{})
// Result: Settings contains {"theme": "dark", "fontSize": 14}, Tags contains ["go", "testing", "automation"]
```

#### Complex Structures
```go
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

type Team struct {
    Lead    User                   `testfill:"unmarshal:{\"name\":\"Alice\",\"age\":30}"`
    Members []User                 `testfill:"unmarshal:[{\"name\":\"Bob\",\"age\":25},{\"name\":\"Carol\",\"age\":28}]"`
    Config  map[string]interface{} `testfill:"unmarshal:{\"maxSize\":10,\"isActive\":true}"`
}

team, _ := testfill.Fill(Team{})
// Result: Fully populated team with JSON data
```

#### Interface{} Fields
```go
type Dynamic struct {
    Data interface{} `testfill:"unmarshal:{\"type\":\"user\",\"id\":123,\"roles\":[\"admin\",\"user\"]}"`
}

dynamic, _ := testfill.Fill(Dynamic{})
// Result: Data contains a map with the JSON structure
```

#### Null Values
```go
type Nullable struct {
    OptionalName *string `testfill:"unmarshal:null"`
    RequiredName *string `testfill:"unmarshal:\"John\""`
}

nullable, _ := testfill.Fill(Nullable{})
// Result: OptionalName is nil, RequiredName points to "John"
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

- Struct value maps only support string keys (use `key:fill`, `key:variant_name`, or `variants:name1,name2,name3` syntax)
- Struct slices and maps support either `fill:count`/`key:fill` for identical instances or variants for different field values
- Interface fields cannot be filled
- Channels and functions are not supported
- Unexported fields are ignored

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
