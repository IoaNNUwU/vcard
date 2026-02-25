package vcard

import (
	"fmt"
	"testing"
)

// Map Marshaling Tests

type telRequiredSchema struct {
	N   string `vCard:"required"`
	TEL string `vCard:"required"`
}

func TestMapMissingRequiredKey(t *testing.T) {

	m := map[string]string{
		"N":  ":Alex;",
		"FN": ":Alex FullName",
		// missing field TEL
	}

	b, err := MarshalSchema(m, SchemaFor[telRequiredSchema]("4.0"))

	assertErrIs(t, err, ErrVCard, "does not contain field \"TEL\"")
	assertSlicesEq(t, b, []byte{})
}

type empty struct{}

var EmptySchema = SchemaFor[empty]("4.0")

func TestEmptyMap(t *testing.T) {

	m := map[string]string{}

	b, _ := MarshalSchema(m, EmptySchema)

	exp := `BEGIN:VCARD
VERSION:4.0
END:VCARD
`
	assertStringsEq(t, string(b), crlfy(exp))
}

func TestCrlfy(t *testing.T) {
	exp := "Hello\r\nWorld\r\n"

	s1 := "Hello\nWorld"
	assertStringLinesEq(t, crlfy(s1), exp)
	assertStringsEq(t, crlfy(s1), exp)

	s2 := `Hello
World`
	assertStringLinesEq(t, crlfy(s2), exp)
	assertStringsEq(t, crlfy(s2), exp)
}

func TestMapCrlf(t *testing.T) {
	m := map[string]string{
		"N":    ":Alex",
		"FN":   ":Alex FullName",
		"NAME": ":Alex Name Hello",
	}

	b, _ := Marshal(m)

	exp := "BEGIN:VCARD\r\nVERSION:4.0\r\nN:Alex\r\nFN:Alex FullName\r\nNAME:Alex Name Hello\r\nEND:VCARD\r\n"
	assertStringLinesEq(t, string(b), exp)
}

func TestMapStringString(t *testing.T) {

	m := map[string]string{
		"N":    ":Alex",
		"FN":   ":Alex FullName",
		"NAME": ":Alex Name Hello",
	}

	b, _ := Marshal(m)

	exp := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

func TestMapMoreFields(t *testing.T) {

	m := map[string]string{
		"N":     ":Alex",
		"FN":    ":Alex FullName",
		"NAME":  ":Alex Name Hello",
		"HELLO": ":World", // Should not be encoded
	}

	b, _ := Marshal(m)

	exp := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

func TestMapStringStringSmart(t *testing.T) {
	m := map[string]string{
		"N":    "Alex",
		"FN":   "Alex FullName",
		"NAME": "Alex Name Hello",
	}

	b, _ := Marshal(m)

	exp := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

type marshalVCardImpl struct {
	s string
}

func (s marshalVCardImpl) MarshalVCardField() ([]byte, error) {
	return fmt.Appendf(nil, ";;%s;;", s.s), nil
}

func TestMapWithValueCustomMarshaler(t *testing.T) {

	m := map[string]marshalVCardImpl{
		"N":    {"Alex"},
		"FN":   {"Alex FullName"},
		"NAME": {"Alex Name Hello"},
	}

	b, _ := Marshal(m)

	exp := `BEGIN:VCARD
VERSION:4.0
N;;Alex;;
FN;;Alex FullName;;
NAME;;Alex Name Hello;;
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

type marshalVCardPtrImpl struct {
	s string
}

func (s *marshalVCardPtrImpl) MarshalVCardField() ([]byte, error) {
	return fmt.Appendf(nil, ";;%s;;", s.s), nil
}

func TestMapWithPtrCustomMarshaler(t *testing.T) {

	m := map[string]marshalVCardPtrImpl{
		"N":    {"Alex"},
		"FN":   {"Alex FullName"},
		"NAME": {"Alex Name Hello"},
	}

	b, _ := Marshal(m)

	exp := `BEGIN:VCARD
VERSION:4.0
N;;Alex;;
FN;;Alex FullName;;
NAME;;Alex Name Hello;;
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

func TestInterfaceWithCustomMarshaler(t *testing.T) {

	m := map[string]VCardFieldMarshaler{
		"N":    marshalVCardImpl{"Alex"},
		"FN":   marshalVCardImpl{"Alex FullName"},
		"NAME": marshalVCardImpl{"Alex Name Hello"},
	}

	b, _ := Marshal(m)

	exp := `BEGIN:VCARD
VERSION:4.0
N;;Alex;;
FN;;Alex FullName;;
NAME;;Alex Name Hello;;
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

func TestAnyWithCustomMarshaler(t *testing.T) {

	m := map[string]any{
		"N":    marshalVCardImpl{"Alex"},
		"FN":   marshalVCardImpl{"Alex FullName"},
		"NAME": marshalVCardImpl{"Alex Name Hello"},
	}

	b, _ := Marshal(m)

	exp := `BEGIN:VCARD
VERSION:4.0
N;;Alex;;
FN;;Alex FullName;;
NAME;;Alex Name Hello;;
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

func TestMapWithUnsupportedKey(t *testing.T) {

	m := map[int]string{
		1: ":Alex",
		2: ":Alex FullName",
	}

	b, err := Marshal(m)

	assertErrIs(t, err, ErrVCard, "type int is not supported as a map key")
	assertSlicesEq(t, b, []byte{})
}

type notSerializable struct {
	s string
}

func TestValueStructDoesNotImplementMarshaler(t *testing.T) {

	m := map[string]notSerializable{
		"N":    {"Alex"},
		"FN":   {"Alex FullName"},
		"NAME": {"Alex Name Hello"},
	}

	b, err := Marshal(m)

	assertErrIs(t, err, ErrVCard, "does not implement VCardFieldMarshaler")
	assertSlicesEq(t, b, []byte{})
}

func TestMapAnyValueDoesNotImplementMarshaler(t *testing.T) {

	m := map[string]any{
		"N":    notSerializable{"Alex"},
		"FN":   notSerializable{"Alex FullName"},
		"NAME": notSerializable{"Alex Name Hello"},
	}

	b, err := Marshal(m)

	assertErrIs(t, err, ErrVCard, "does not implement VCardFieldMarshaler")
	assertSlicesEq(t, b, []byte{})
}

func TestUnsupportedTypeAsMapValue(t *testing.T) {

	m := map[string]int{
		"N":    10,
		"FN":   11,
		"NAME": 16,
	}

	b, err := Marshal(m)

	assertErrIs(t, err, ErrVCard, "type int is not supported as a map value")
	assertSlicesEq(t, b, []byte{})
}

func TestMarshalEmptySlice(t *testing.T) {
	sl := []map[string]string{}
	b, _ := Marshal(sl)
	assertSlicesEq(t, b, []byte{})
}

func TestMarshalSliceOfMaps(t *testing.T) {

	sl := []map[string]string{
		{
			"N":    ":Alex 1",
			"FN":   ":Alex FullName 1",
			"NAME": ":Alex Name Hello 1",
		},
		{
			"N":    ":Alex 2",
			"FN":   ":Alex FullName 2",
			"NAME": ":Alex Name Hello 2",
		},
		{
			"N":    ":Alex 3",
			"FN":   ":Alex FullName 3",
			"NAME": ":Alex Name Hello 3",
		},
	}

	b, _ := Marshal(sl)

	exp := `BEGIN:VCARD
VERSION:4.0
N:Alex 1
FN:Alex FullName 1
NAME:Alex Name Hello 1
END:VCARD
BEGIN:VCARD
VERSION:4.0
N:Alex 2
FN:Alex FullName 2
NAME:Alex Name Hello 2
END:VCARD
BEGIN:VCARD
VERSION:4.0
N:Alex 3
FN:Alex FullName 3
NAME:Alex Name Hello 3
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

// Struct Marshaling Tests

type missingFieldImpl struct {
	N  string
	FN string
	// missing field TEL
}

func TestStructMissingRequiredField(t *testing.T) {
	stru := missingFieldImpl{
		N:  ":Alex",
		FN: ":Alex FullName",
	}

	b, err := MarshalSchema(stru, SchemaFor[telRequiredSchema]("4.0"))

	assertErrIs(t, err, ErrVCard, "does not contain field \"TEL\"")
	assertSlicesEq(t, b, []byte{})
}

func TestEmptyStruct(t *testing.T) {

	s := empty{}

	b, _ := MarshalSchema(s, EmptySchema)

	exp := `BEGIN:VCARD
VERSION:4.0
END:VCARD
`
	assertStringsEq(t, string(b), crlfy(exp))
}

type stringUser struct {
	N    string
	FN   string
	NAME string
}

func TestStructStringFields(t *testing.T) {

	s := stringUser{
		N:    ":Alex",
		FN:   ":Alex FullName",
		NAME: ":Alex Name Hello",
	}

	b, _ := Marshal(s)

	exp := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

type moreStringUser struct {
	N     string
	FN    string
	NAME  string
	HELLO string
}

func TestStructMoreFields(t *testing.T) {

	s := moreStringUser{
		N:     ":Alex",
		FN:    ":Alex FullName",
		NAME:  ":Alex Name Hello",
		HELLO: ":World",
	}

	b, _ := Marshal(s)

	exp := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

func TestStructStringFieldsSmart(t *testing.T) {

	s := stringUser{
		N:    "Alex",
		FN:   "Alex FullName",
		NAME: "Alex Name Hello",
	}

	b, _ := Marshal(s)

	exp := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

type customMarshalerUser struct {
	N    marshalVCardImpl
	FN   marshalVCardImpl
	NAME marshalVCardImpl
}

func TestStructCustomFields(t *testing.T) {

	s := customMarshalerUser{
		N:    marshalVCardImpl{"Alex"},
		FN:   marshalVCardImpl{"Alex FullName"},
		NAME: marshalVCardImpl{"Alex Name Hello"},
	}

	b, _ := Marshal(s)

	exp := `BEGIN:VCARD
VERSION:4.0
N;;Alex;;
FN;;Alex FullName;;
NAME;;Alex Name Hello;;
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

type customMarshalerPtrUser struct {
	N    *marshalVCardImpl
	FN   *marshalVCardImpl
	NAME *marshalVCardImpl
}

func TestStructCustomPtrFields(t *testing.T) {

	s := customMarshalerPtrUser{
		N:    &marshalVCardImpl{"Alex"},
		FN:   &marshalVCardImpl{"Alex FullName"},
		NAME: &marshalVCardImpl{"Alex Name Hello"},
	}

	b, _ := Marshal(s)

	exp := `BEGIN:VCARD
VERSION:4.0
N;;Alex;;
FN;;Alex FullName;;
NAME;;Alex Name Hello;;
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

type customMarshalerInterfaceUser struct {
	N    VCardFieldMarshaler
	FN   VCardFieldMarshaler
	NAME VCardFieldMarshaler
}

func TestStructInterfaceMarshalerFields(t *testing.T) {

	s := customMarshalerInterfaceUser{
		N:    marshalVCardImpl{"Alex"},
		FN:   marshalVCardImpl{"Alex FullName"},
		NAME: marshalVCardImpl{"Alex Name Hello"},
	}

	b, _ := Marshal(s)

	exp := `BEGIN:VCARD
VERSION:4.0
N;;Alex;;
FN;;Alex FullName;;
NAME;;Alex Name Hello;;
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

type anyUser struct {
	N    any
	FN   any
	NAME any
}

func TestStructAnyMarshalerFields(t *testing.T) {

	s := anyUser{
		N:    marshalVCardImpl{"Alex"},
		FN:   marshalVCardImpl{"Alex FullName"},
		NAME: marshalVCardImpl{"Alex Name Hello"},
	}

	b, _ := Marshal(s)

	exp := `BEGIN:VCARD
VERSION:4.0
N;;Alex;;
FN;;Alex FullName;;
NAME;;Alex Name Hello;;
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

type tagsUser struct {
	Name        string `vCard:"N"`
	FullName    string `vCard:"FN"`
	Description string `vCard:"NAME"`
}

func TestStructHasRenameTags(t *testing.T) {

	s := tagsUser{
		Name:        ":Alex",
		FullName:    ":Alex FullName",
		Description: ":Alex Name Hello",
	}

	b, _ := Marshal(s)

	exp := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}

func TestStructAnyDoesNotImplementMarshaler(t *testing.T) {

	s := anyUser{
		N:    notSerializable{"Alex"},
		FN:   notSerializable{"Alex FullName"},
		NAME: notSerializable{"Alex Name Hello"},
	}

	b, err := Marshal(s)

	assertErrIs(t, err, ErrVCard, "does not implement VCardFieldMarshaler")
	assertSlicesEq(t, b, []byte{})
}

type unsupportedUser struct {
	N    notSerializable
	FN   notSerializable
	NAME notSerializable
}

func TestStructDoesNotImplementMarshaler(t *testing.T) {

	s := unsupportedUser{
		N:    notSerializable{":Alex"},
		FN:   notSerializable{":Alex FullName"},
		NAME: notSerializable{":Alex Name Hello"},
	}

	b, err := Marshal(s)

	assertErrIs(t, err, ErrVCard, "does not implement VCardFieldMarshaler")
	assertSlicesEq(t, b, []byte{})
}

func TestMarshalSliceOfStructs(t *testing.T) {

	sl := []customMarshalerUser{
		{
			N:    marshalVCardImpl{"Alex 1"},
			FN:   marshalVCardImpl{"Alex FullName 1"},
			NAME: marshalVCardImpl{"Alex Name Hello 1"},
		},
		{
			N:    marshalVCardImpl{"Alex 2"},
			FN:   marshalVCardImpl{"Alex FullName 2"},
			NAME: marshalVCardImpl{"Alex Name Hello 2"},
		},
		{
			N:    marshalVCardImpl{"Alex 3"},
			FN:   marshalVCardImpl{"Alex FullName 3"},
			NAME: marshalVCardImpl{"Alex Name Hello 3"},
		},
	}

	b, _ := Marshal(sl)

	exp := `BEGIN:VCARD
VERSION:4.0
N;;Alex 1;;
FN;;Alex FullName 1;;
NAME;;Alex Name Hello 1;;
END:VCARD
BEGIN:VCARD
VERSION:4.0
N;;Alex 2;;
FN;;Alex FullName 2;;
NAME;;Alex Name Hello 2;;
END:VCARD
BEGIN:VCARD
VERSION:4.0
N;;Alex 3;;
FN;;Alex FullName 3;;
NAME;;Alex Name Hello 3;;
END:VCARD
`
	assertStringLinesEq(t, string(b), crlfy(exp))
}
