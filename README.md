# testfill

A Go library for filling struct fields with test data using struct tags.

```go
// Define test data in struct tags
type User struct {
    Name    string  `testfill:"John Doe"`
    Age     int     `testfill:"30"`
    Email   string  `testfill:"john@example.com"`
    Address Address `testfill:"fill"`
}

// Fill struct with tagged values
user, _ := testfill.Fill(User{})
```

## Installation

```bash
go get github.com/fabioelizandro/testfill
```

## Basic Usage

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

Only zero-valued fields are filled. Existing values are preserved:

```go
config, _ := testfill.Fill(Config{Port: 3000}) // Port already set
// Result: {Host:localhost Port:3000} // Host filled, Port preserved
```

## Nested Structs

```go
type User struct {
    Name    string  `testfill:"Jane Doe"`
    Address Address `testfill:"fill"`
}

user, _ := testfill.Fill(User{})
// Recursively fills Address fields
```

## Collections

```go
type TestData struct {
    Tags     []string `testfill:"go,testing,automation"`
    Users    []User   `testfill:"fill:3"`
    Variants []User   `testfill:"variants:admin,user,guest"`
}
```

## Variants

```go
type User struct {
    Name string `testfill:"John" testfill_admin:"Jane" testfill_guest:"Bob"`
    Role string `testfill:"user" testfill_admin:"admin" testfill_guest:"guest"`
}

adminUser, _ := testfill.FillWithVariant(User{}, "admin")
// Result: {Name:Jane Role:admin}
```

## Factory Functions

```go
testfill.RegisterFactory("uuid", func() string {
    return uuid.New().String()
})

type Document struct {
    ID string `testfill:"factory:uuid"`
}
```

## JSON Unmarshaling

```go
type Config struct {
    Settings map[string]interface{} `testfill:"unmarshal:{\"theme\":\"dark\"}"`
    Tags     []string              `testfill:"unmarshal:[\"go\",\"testing\"]"`
}
```

## API

```go
// Fill struct with default values
user, err := testfill.Fill(User{})

// Fill with specific variant
adminUser, err := testfill.FillWithVariant(User{}, "admin")

// Panic versions
user := testfill.MustFill(User{})
adminUser := testfill.MustFillWithVariant(User{}, "admin")
```

## Tag Syntax

- `testfill:"value"` - Basic value
- `testfill:"fill"` - Fill nested struct
- `testfill:"val1,val2,val3"` - Slice values  
- `testfill:"fill:3"` - Generate 3 structs
- `testfill:"variants:admin,user"` - Use variants
- `testfill:"factory:name:arg1:arg2"` - Factory function
- `testfill:"unmarshal:{\"key\":\"value\"}"` - JSON data

## Supported Types

**Supported:** primitives, slices, maps, pointers, nested structs, time.Time  
**Not supported:** interfaces, channels, functions, unexported fields

## Error Handling

```go
_, err := testfill.Fill(Invalid{})
// Returns descriptive error messages for type conversion failures
```
