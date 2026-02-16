package vcard

import (
	"fmt"
	"testing"
)

// Map Marshaling Tests

type TelRequiredSchema struct {
	N   string `vCard:"required"`
	TEL string `vCard:"required"`
}

func TestMapMissingRequiredKey(t *testing.T) {

	m := map[string]string{
		"N":  ":Alex;",
		"FN": ":Alex FullName",
		// missing field TEL
	}

	b, err := MarshalSchema(m, SchemaFor[TelRequiredSchema]("4.0"))

	assertErrIs(t, err, ErrVCard, "does not contain field \"TEL\"")
	assertSlicesEq(t, b, []byte{})
}

type Empty struct{}

var EmptySchema = SchemaFor[Empty]("4.0")

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

type MarshalVCardImpl struct {
	s string
}

func (s MarshalVCardImpl) MarshalVCardField() ([]byte, error) {
	return fmt.Appendf(nil, ";;%s;;", s.s), nil
}

func TestMapWithValueCustomMarshaler(t *testing.T) {

	m := map[string]MarshalVCardImpl{
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
		"N":    MarshalVCardImpl{"Alex"},
		"FN":   MarshalVCardImpl{"Alex FullName"},
		"NAME": MarshalVCardImpl{"Alex Name Hello"},
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
		"N":    MarshalVCardImpl{"Alex"},
		"FN":   MarshalVCardImpl{"Alex FullName"},
		"NAME": MarshalVCardImpl{"Alex Name Hello"},
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

type NotMarshaler struct {
	s string
}

func TestValueStructDoesNotImplementMarshaler(t *testing.T) {

	m := map[string]NotMarshaler{
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
		"N":    NotMarshaler{"Alex"},
		"FN":   NotMarshaler{"Alex FullName"},
		"NAME": NotMarshaler{"Alex Name Hello"},
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

type MissingFieldImpl struct {
	N  string
	FN string
	// missing field TEL
}

func TestStructMissingRequiredField(t *testing.T) {
	stru := MissingFieldImpl{
		N:  ":Alex",
		FN: ":Alex FullName",
	}

	b, err := MarshalSchema(stru, SchemaFor[TelRequiredSchema]("4.0"))

	assertErrIs(t, err, ErrVCard, "does not contain field \"TEL\"")
	assertSlicesEq(t, b, []byte{})
}

func TestEmptyStruct(t *testing.T) {

	s := Empty{}

	b, _ := MarshalSchema(s, EmptySchema)

	exp := `BEGIN:VCARD
VERSION:4.0
END:VCARD
`
	assertStringsEq(t, string(b), crlfy(exp))
}

type StringUser struct {
	N    string
	FN   string
	NAME string
}

func TestStructStringFields(t *testing.T) {

	s := StringUser{
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

type MoreStringUser struct {
	N     string
	FN    string
	NAME  string
	HELLO string
}

func TestStructMoreFields(t *testing.T) {

	s := MoreStringUser{
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

	s := StringUser{
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

type CustomMarshalerUser struct {
	N    MarshalVCardImpl
	FN   MarshalVCardImpl
	NAME MarshalVCardImpl
}

func TestStructCustomFields(t *testing.T) {

	s := CustomMarshalerUser{
		N:    MarshalVCardImpl{"Alex"},
		FN:   MarshalVCardImpl{"Alex FullName"},
		NAME: MarshalVCardImpl{"Alex Name Hello"},
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

type CustomMarshalerInterfaceUser struct {
	N    VCardFieldMarshaler
	FN   VCardFieldMarshaler
	NAME VCardFieldMarshaler
}

func TestStructInterfaceMarshalerFields(t *testing.T) {

	s := CustomMarshalerInterfaceUser{
		N:    MarshalVCardImpl{"Alex"},
		FN:   MarshalVCardImpl{"Alex FullName"},
		NAME: MarshalVCardImpl{"Alex Name Hello"},
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

type AnyUser struct {
	N    any
	FN   any
	NAME any
}

func TestStructAnyMarshalerFields(t *testing.T) {

	s := AnyUser{
		N:    MarshalVCardImpl{"Alex"},
		FN:   MarshalVCardImpl{"Alex FullName"},
		NAME: MarshalVCardImpl{"Alex Name Hello"},
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

type TagsUser struct {
	name        string `vCard:"N"`
	fullName    string `vCard:"FN"`
	description string `vCard:"NAME"`
}

func TestStructHasRenameTags(t *testing.T) {

	s := TagsUser{
		name:        ":Alex",
		fullName:    ":Alex FullName",
		description: ":Alex Name Hello",
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

	s := AnyUser{
		N:    NotMarshaler{"Alex"},
		FN:   NotMarshaler{"Alex FullName"},
		NAME: NotMarshaler{"Alex Name Hello"},
	}

	b, err := Marshal(s)

	assertErrIs(t, err, ErrVCard, "does not implement VCardFieldMarshaler")
	assertSlicesEq(t, b, []byte{})
}

type UnsupportedUser struct {
	N    NotMarshaler
	FN   NotMarshaler
	NAME NotMarshaler
}

func TestStructDoesNotImplementMarshaler(t *testing.T) {

	s := UnsupportedUser{
		N:    NotMarshaler{":Alex"},
		FN:   NotMarshaler{":Alex FullName"},
		NAME: NotMarshaler{":Alex Name Hello"},
	}

	b, err := Marshal(s)

	assertErrIs(t, err, ErrVCard, "does not implement VCardFieldMarshaler")
	assertSlicesEq(t, b, []byte{})
}

func TestMarshalSliceOfStructs(t *testing.T) {

	sl := []CustomMarshalerUser{
		{
			N:    MarshalVCardImpl{"Alex 1"},
			FN:   MarshalVCardImpl{"Alex FullName 1"},
			NAME: MarshalVCardImpl{"Alex Name Hello 1"},
		},
		{
			N:    MarshalVCardImpl{"Alex 2"},
			FN:   MarshalVCardImpl{"Alex FullName 2"},
			NAME: MarshalVCardImpl{"Alex Name Hello 2"},
		},
		{
			N:    MarshalVCardImpl{"Alex 3"},
			FN:   MarshalVCardImpl{"Alex FullName 3"},
			NAME: MarshalVCardImpl{"Alex Name Hello 3"},
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
