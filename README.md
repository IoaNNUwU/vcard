# VCard support in Golang

This package provides a simple reflection-based API to encode and decode [`vCard` documents](https://en.wikipedia.org/wiki/VCard).

- API is based on `json` package from standard library.
- `reflect` is used to automatically marshal/unmarshal user-defined types and custom fields.
- supports field renaming using `tags`.
- supports custom [`Schema` definition](#schema-definition) with set of built-in schemas for vCard `4.0`, `3.0`, `2.1`.

This library is still **WIP** with main concern being performance. Contributions and issues are welcome.

## Basic Usage

Read vCard document using `io` module:
```go
file, _ := os.Open("my_vcard_file.vcf")
data, _ := io.ReadAll(file)
```
### [Decode](#custom-decoding-logic) this byte slice into a Go value by providing it by a pointer to `vcard.Unmarshal` function. 

Decoding is only possible into a `struct`, `map` or a `slice`.
```go
// Unmarshal single (or first) vCard document into a map
m := make(map[string]string)
err := vcard.Unmarshal(data, &m)
assert(errors.Is(err, vcard.ErrLeftoverTokens))

// Unmarshal multiple vCard documents into a slice of maps
slice := make([]map[string]string, 0)
err := vcard.Unmarshal(data, &slice)
assert(err == nil)
```
Decode into a user-defined `struct`:
```go
type MyUser struct {
    FullName string `vCard:"FN"`
    Name     string `vCard:"N"`
    TEL      string
}
// Unmarshal single (or first) vCard document into a struct
u := MyUser{}
err := vcard.Unmarshal(data, &u)
assert(errors.Is(err, vcard.ErrLeftoverTokens))

// Unmarshal multiple vCard documents into a slice of structs
slice := make([]MyUser, 0)
err := vcard.Unmarshal(data, &slice)
assert(err == nil)
```

- See [custom decoding logic](#custom-decoding-logic) if you want to use non-`string` fields in your `struct`.

### [Encode](#custom-encoding-logic) a Go value using `vcard.Marshal` function. 
Encoding is only possible with `struct`, `map` or a `slice`.
```go
// Marshal map m into a byte slice
var m map[string]string
bytes, err := vcard.Marshal(m)

// Marshal struct into a byte slice
var u MyUser
bytes, err := vcard.Marshal(u)

// Marshal a slice of maps into bytes
var slice []map[string]string
bytes, err := vcard.Marshal(slice)

// Marshal a slice of structs into bytes
var slice []MyUser
bytes, err := vcard.Marshal(slice)
```

- See [custom encoding logic](#custom-encoding-logic) if you want to use non-`string` fields in your struct.

## Custom encoding logic

`VCardFieldMarshaler` interface defines custom encoding logic for a single field inside VCard document.
```go
type VCardFieldMarshaler interface {
	MarshalVCardField() ([]byte, error)
}
```
Let's encode line `TEL;TYPE=CELL:(123) 555-5832` in this document:
```
BEGIN:VCARD
VERSION:4.0
FN:Alex
TEL;TYPE=CELL:(123) 555-5832
END:VCARD
```
- `TEL` is a key
- `;TYPE=CELL:(123) 555-5832` is a value we need to marshal

```go
type Tel struct {
    typ string
    tel string
}

func (t *Tel) MarshalVCardField() ([]byte, error) {
    // final result is ";TYPE=CELL:(123) 555-5832"
    return fmt.Sprintf(";TYPE=%s:%s", t.typ, t.tel), nil
}

// This struct will use built-in marshaling for FN field and 
// VCardFieldMarshaler implementation for TEL field of type Tel
type MyCustomUser struct {
    FN  string
    TEL Tel
}
```

## Custom decoding logic

`VCardFieldUnmarshaler` interface defines custom decoding logic for a single field inside VCard document.
```go
type VCardFieldUnmarshaler interface {
	UnmarshalVCardField(data []byte) error
}
```
Let's decode line `TEL;TYPE=CELL:(123) 555-5832` in this document:
```
BEGIN:VCARD
VERSION:4.0
FN:Alex
TEL;TYPE=CELL:(123) 555-5832
END:VCARD
```
- `TEL` is a key
- `;TYPE=CELL:(123) 555-5832` is a value we need to unmarshal

```go
type Tel struct {
    typ string
    tel string
}

func (t *Tel) UnmarshalVCardField(data []byte) error {
    // data has format ";TYPE=CELL:(123) 555-5832"
    s := string(data)

    // sl[0] is ";TYPE=CELL"
    // sl[1] is "(123) 555-5832"
    sl := strings.Split(s, ":")
    if len(sl) != 2 {
        return errors.New("Wrong field format")
    }
    if strings.Contains(sl[0], "VOICE") {
        t.typ = "VOICE"
    } else {
        t.typ = "CELL"
    }
    t.tel = sl[1]

    return nil
}

// This struct will use built-in unmarshaling for FN field and
// VCardFieldUnmarshaler implementation for TEL field of type Tel
type MyCustomUser struct {
    FN  string
    TEL Tel
}
```

## Schema definition

Both `Encoder` and `Decoder` support defining custom set of schemas for your needs including:
- Set of fields that need to be serialized. Encoder/Decoder ignore fields outside of this list.
- Set of `required` fields. Decoder returns `ErrParsing` if `required` field is missing from vCard document. Encoder returns `ErrVCard` if struct/map is missing `required` field.
- You could provide set of schemas with different `VERSION`s to `Decoder` and use single `Decoder` instance to deal with multiple vCard records of different versions in single file.

This is useful in case of specific requirements, e.g.:

- `TEL` field is not required by standard, let's use custom schema to ensure it's presence.

```go
type CustomSchemaV4 struct {
    N     string `vCard:"required"`
    TEL   string `vCard:"required"`
    EMAIL string
}

var MySchemaV4 := SchemaFor[CustomSchemaV4]("4.0")

m := map[string]string {
    "N":     "Alex",
    "EMAIL": "alex@example.com"
    // TEL key is missing
}

b, err := MarshalSchema(m, MySchemaV4)
// err map missing key TEL required by the schema

m := map[string]string {
    "N":     "Alex",
    "EMAIL": "alex@example.com",
    "TEL":   "333 555",

    "HELLO": "World", // HELLO key is not in the schema and will be ignored
}

b, err := MarshalSchema(m, MySchemaV4)

// resulting b is:
//
// BEGIN:VCARD
// VERSION:4.0
// N:Alex
// EMAIL:alex@example.com
// TEL:333 555
// END:VCARD
```
