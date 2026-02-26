# VCard support in Golang

This library is still **WIP** with main concern being performance. Contributions and issues are welcome.

See [Examples](./examples/) for example usage.

- See [Schema definition](#schema-definition) or use built-in set of schemas via `Marshal`/`Unmarshal`
- Create a struct or a map with names from the schema (e.g. `FN`, `N`, `TEL` are valid field/key names).
- Use `string` or a struct that implements [`VCardFieldMarshaler`](#custom-encoding-logic)/[`VCardFieldUnmarshaler`](#custom-decoding-logic) as your fields type.
- Use `vcard.Marshal(v any) ([]byte, error)` for encoding.
- Use `vcard.Unmarshal(data []byte, v any) error` for decoding.

### Using structs

You can tag field in a struct with `vCard:"FN"`, `vCard:"N"`, `vCard:"TEL"` to give them alternative name
```go
type MyUser struct {
    Name string `vCard:"FN"`
    Tel  string `vCard:"TEL"`
}
```
- Fields have to either be a `string` or struct that implements `VCardFieldMarshaler`/`VCardFieldUnmarshaler`
- Encoding is also allowed for `interface` fields e.g. `any` or `VCardFieldMarshaler` as a type.

## Using maps

- Map key has to be a `string`
- Map value has to either be a `string` or struct that implements `VCardFieldMarshaler`/`VCardFieldUnmarshaler`

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
