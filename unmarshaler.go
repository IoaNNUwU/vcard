package vcard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Deserializes a vCard document into a Go value.
//
// v has to be a pointer to a slice, struct or a map.
//
// Short-hand for Decoder.Decode() with default set of schemas.
func Unmarshal(data []byte, v any) error {
	r := bytes.NewReader(data)
	dec := NewDecoder(r, DefaultSchemas)
	return dec.Decode(v)
}

type Decoder struct {
	r io.Reader

	// maps version string to schema
	schemas map[string]Schema
}

// Creates new Decoder that reads from r using provided schemas.
//
// panics if schemas slice has multiple schemas with same version.
// if schemas slice is empty
func NewDecoder(r io.Reader, schemas []Schema) *Decoder {
	m := make(map[string]Schema)

	for _, s := range schemas {
		_, exists := m[s.version]
		if exists {
			panic(fmt.Sprintf("vCard: cannot create a Decoder of multiple schemas with same version %s", s.version))
		}
		m[s.version] = s
	}

	return &Decoder{
		r:       r,
		schemas: m,
	}
}

var ErrParse = errors.New("vCard: parsing error in Decoder")

// Decodes a vCard document into pointer v using provided schema.
//
// Returns [ErrParse] in case of a malformed bytes from writer.
//
// v has to be a pointer to a struct, map or a slice.
func (d *Decoder) Decode(v any) error {
	b, err := io.ReadAll(d.r)
	if err != nil {
		return fmt.Errorf("vCard: unable to read: %w", err)
	}
	maybePtr := reflect.ValueOf(v)

	if maybePtr.Kind() != reflect.Pointer {
		return fmt.Errorf("vCard: decoding is only possible into a pointer, not %s", maybePtr.Kind())
	}
	if maybePtr.IsNil() {
		return errors.New("vCard: decoding is only possible into a not nil pointer")
	}
	value := maybePtr.Elem()

	return d.decode(string(b), value)
}

func (d *Decoder) decode(s string, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Map:
		return d.decodeMap(s, v)
	case reflect.Struct:
		return d.decodeStruct(s, v)
	case reflect.Slice:
		return d.decodeSlice(s, v)
	case reflect.Array:
		return d.decodeArray(s, v)
	default:
		return fmt.Errorf("vCard: unable to decode into %s type. Use struct, map or a slice", v.Type())
	}
}

func (d *Decoder) decodeMap(s string, v reflect.Value) error {

	return nil
}

func (d *Decoder) decodeStruct(s string, v reflect.Value) error {

	s, err := d.decodeRecordHeader(s)
	if err != nil {
		return err
	}

	s, err = d.decodeRecordFooter(s)
	if err != nil {
		return err
	}

	return nil
}

func (d *Decoder) decodeSlice(s string, v reflect.Value) error {
	panic("TODO: decodeSlice")
}

func (d *Decoder) decodeArray(s string, v reflect.Value) error {
	panic("TODO: decodeArray")
}

func (d *Decoder) decodeRecordHeader(s string) (string, error) {
	if s == "" {
		return s, fmt.Errorf("%w: %w", ErrParse, io.ErrUnexpectedEOF)
	}
	expectedHeader := "BEGIN:VCARD\n"
	for line := range strings.Lines(s) {
		if line != expectedHeader {
			return s, fmt.Errorf("%w: expected %q but found %q", ErrParse, expectedHeader, line)
		}
		break
	}
	return s[len(expectedHeader):], nil
}

func (d *Decoder) decodeRecordFooter(s string) (string, error) {
	if s == "" {
		return s, fmt.Errorf("%w: %w", ErrParse, io.ErrUnexpectedEOF)
	}
	expectedFooter := "END:VCARD\n"
	for line := range strings.Lines(s) {
		if line != expectedFooter && line != expectedFooter[:len(expectedFooter)-1] {
			return s, fmt.Errorf("%w: expected %q but found %q", ErrParse, expectedFooter, line)
		}
		break
	}
	return s[len(expectedFooter)-1:], nil
}

// Implemented by fields that need custom Unmarshaling logic.
//
// Note that this interface defines a way to unmarshal single field.
// e.g. TEL field has custom type Tel:
//
//	    type MySchemaV4 struct {
//			FN  string `vCard:"required"`
//			TEL Tel    `vCard:"required"`
//		}
//
//		type Tel struct {
//			typ string
//			tel string
//		}
//
//		func (t Tel) UnmarshalVCardField(data []byte) error {
//			// data has a form of ";TYPE=CELL:(123) 555-5832"
//			s := string(data)
//
//			sl := strings.Split(s, ":")
//			if len(sl) != 2 {
//				return errors.New("Unable to unmarshal")
//			}
//
//			if strings.Contains(sl[0], "VOICE") {
//				t.typ = "VOICE"
//			} else {
//				t.typ = "CELL"
//			}
//			t.tel = sl[1]
//
//			return nil
//		}
type VCardFieldUnmarshaler interface {
	UnmarshalVCardField(data []byte) error
}
