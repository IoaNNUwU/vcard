package vcard

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode"
)

// Deserializes a vCard document into a Go value using default set of [Schema]s.
//
// v has to be a pointer to a slice, struct or a map.
func Unmarshal(data []byte, v any) error {
	return UnmarshalSchema(data, v, DefaultSchemas)
}

// Deserializes a vCard document into a Go value using provided set of [Schema]s.
//
// v has to be a pointer to a slice, struct or a map.
func UnmarshalSchema(data []byte, v any, schemas []Schema) error {
	r := bytes.NewReader(data)
	dec := NewDecoder(r, schemas)
	return dec.Decode(v)
}

// Reads a vCard document from an input stream.
type Decoder struct {
	r io.Reader

	// maps version string to schema
	schemas map[string]Schema

	smartStrings bool

	// TODO: Decoder setting to be precise about line formatting
	// e.g. ignore spaces and newline sequence
}

// Creates new Decoder that reads from r using provided schemas.
//
// panics if schemas slice has multiple schemas with same version.
// if schemas slice is empty
func NewDecoder(r io.Reader, schemas []Schema) *Decoder {
	m := make(map[string]Schema)

	for _, s := range schemas {
		_, found := m[s.version]
		if found {
			panic(vCardErrf("cannot create a Decoder of multiple schemas with same version %s", s.version))
		}
		m[s.version] = s
	}

	return &Decoder{
		r:            r,
		schemas:      m,
		smartStrings: true,
	}
}

// Toggles smart string encoding. Enabled by default.
//
// In smart mode, decoder checks at runtime if string starts with `:` (part of KEY:VALUE separator)
// and removes it if neccesary e.g. string fields will contain "Alex" instead of ":Alex".
//
// See [Encoder.SetSmartStrings] for more info.
func (d *Decoder) SetSmartStrings(smartStrings bool) *Decoder {
	d.smartStrings = smartStrings
	return d
}

// Decodes a vCard document into pointer v using provided schema.
//
// Returns [ErrParsing] in case of a malformed vCard document recived from Writer.
//
// v has to be a pointer to a struct, map or a slice.
func (d *Decoder) Decode(v any) error {
	b, err := io.ReadAll(d.r)
	if err != nil {
		return vCardErrf("unable to read: %w", err)
	}
	maybePtr := reflect.ValueOf(v)

	if maybePtr.Kind() != reflect.Pointer {
		return vCardErrf("decoding is only possible into a pointer, not %s", maybePtr.Kind())
	}
	if maybePtr.IsNil() {
		return vCardErrf("decoding is only possible into a not-nil pointer")
	}
	value := maybePtr.Elem()

	_, err = d.decode(string(b), value)
	return err
}

func (d *Decoder) decode(s string, v reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.Map:
		return d.decodeMap(s, v)
	case reflect.Struct:
		return d.decodeStruct(s, v)
	case reflect.Slice:
		return d.decodeSlice(s, v)
	case reflect.Array:
		return d.decodeArray(s, v)
	}
	return s, vCardErrf("unable to decode into %s type. Use struct, map or a slice", v.Type())
}

func (d *Decoder) decodeMap(data string, ma reflect.Value) (string, error) {
	if ma.IsNil() {
		return data, vCardErrf("decoding is only possible into not-nil map")
	}

	s, err := d.decodeRecordHeader(data)
	if err != nil {
		return data, err
	}
	m, schema, s, err := d.decodeVCardFieldsIntoMap(data)
	if err != nil {
		return data, err
	}
	s, err = d.decodeRecordFooter(s)
	if err != nil {
		return data, err
	}

	err = d.fillMap(ma, m, schema)
	if err != nil {
		return data, err
	}

	if len(strings.TrimSpace(s)) != 0 {
		return s, leftTokensErrf("after successfully decoding a struct")
	}

	return s, nil
}

func (d *Decoder) fillMap(ma reflect.Value, m map[string]string, schema Schema) error {

	key := ma.Type().Key()
	if key.Kind() != reflect.String {
		return vCardErrf("unable to decode into a map where key has unsupported type %s. Use string instead", key)
	}

	elem := ma.Type().Elem()

	switch elem.Kind() {
	case reflect.String:
		newMap := make(map[string]string, len(schema.fields))

		for req := range schema.fields {
			v, found := m[req]
			if !found {
				continue
			}
			newMap[req] = v
		}
		ma.Set(reflect.ValueOf(newMap))

	case reflect.Struct:
		if !elem.Implements(reflect.TypeFor[VCardFieldUnmarshaler]()) {
			return vCardErrf("unable to decode into a map where value has type %s that does not implement VCardFieldUnmarshaler", elem)
		}

		for field := range schema.fields {
			v, found := m[field]
			if !found {
				continue
			}

			value := reflect.Zero(elem)
			i := value.Interface().(VCardFieldUnmarshaler)

			err := i.UnmarshalVCardField([]byte(v))
			if err != nil {
				return vCardErrf("error while unmarshaling a value for a key %q: %w", field, err)
			}
			ma.SetMapIndex(reflect.ValueOf(field), value)
		}

	case reflect.Interface:
		for field := range schema.fields {
			v, found := m[field]
			if !found {
				continue
			}

			value := reflect.Zero(elem)
			i, ok := value.Interface().(VCardFieldUnmarshaler)
			if !ok {
				return vCardErrf("unable to decode a value for a map key %q because it has type %s which does not implement VCardFieldUnmarshaler", key, elem)
			}

			err := i.UnmarshalVCardField([]byte(v))
			if err != nil {
				return vCardErrf("error while unmarshaling a value for a key %q: %w", field, err)
			}
			ma.SetMapIndex(reflect.ValueOf(field), value)
		}
	}

	return vCardErrf("unable to decode into a map where value has unsupported type %s. Use string or struct that implements VCardFieldUnmarshaler", key)
}

func (d *Decoder) decodeStruct(data string, struc reflect.Value) (string, error) {

	s, err := d.decodeRecordHeader(data)
	if err != nil {
		return data, err
	}
	m, schema, s, err := d.decodeVCardFieldsIntoMap(data)
	if err != nil {
		return data, err
	}
	s, err = d.decodeRecordFooter(s)
	if err != nil {
		return data, err
	}

	err = d.fillStruct(struc, m, schema)
	if err != nil {
		return data, err
	}

	if len(strings.TrimSpace(s)) != 0 {
		return s, leftTokensErrf("after successfully decoding a struct")
	}

	return s, nil
}

func (d *Decoder) fillStruct(struc reflect.Value, m map[string]string, schema Schema) error {

	for req := range schema.requiredFields {
		matches := false
		for i := range struc.NumField() {

			field := struc.Type().Field(i)
			vCardName := field.Name

			tag := field.Tag.Get("vCard")
			if tag != "" {
				vCardName = tag
			}
			if req == vCardName {
				matches = true
			}
		}

		if !matches {
			return vCardErrf("struct %s does not contain a field %q or field tagged `vCard:\"%s\"` required by the schema", struc.Type(), req, req)
		}
	}

	for i := range struc.NumField() {
		field := struc.Type().Field(i)
		fieldValue := struc.Field(i)

		vCardName := field.Name

		tag := field.Tag.Get("vCard")
		if tag != "" {
			vCardName = tag
		}

		_, found := schema.fields[vCardName]
		if !found {
			continue
		}
		serField, found := m[vCardName]
		if !found {
			continue
		}

		// Everything is alright, we need to decode this field into v
		if !fieldValue.CanSet() {
			return vCardErrf("unable to set a field %q of struct %s for unexpected reason", field.Name, fieldValue.Type())
		}
		taggedMsg := ""
		if tag != "" {
			taggedMsg = fmt.Sprintf("tagged `vCard:\"%s\"` ", tag)
		}

		switch field.Type.Kind() {
		case reflect.String:
			if !d.smartStrings {
				fieldValue.SetString(serField)
			} else {
				if serField[0] == ':' {
					fieldValue.SetString(serField[1:])
				} else {
					fieldValue.SetString(serField)
				}
			}
		case reflect.Struct, reflect.Interface:
			v, ok := fieldValue.Interface().(VCardFieldUnmarshaler)
			if !ok {
				return vCardErrf("field %q %sof type %s has type %s which does not implement VCardFieldUnmarshaler", field.Name, taggedMsg, struc.Type(), fieldValue.Type())
			}
			err := v.UnmarshalVCardField([]byte(serField))
			if err != nil {
				return vCardErrf("error during unmarshaling field %q %sof struct %s: %w", field.Name, taggedMsg, struc.Type(), err)
			}
		default:
			return vCardErrf("field %q %sof type %shas unsupported type %s. Use string or struct that implements VCardFieldUnmarshaler", field.Name, taggedMsg, struc.Type(), field.Type)
		}
	}

	return nil
}

func (d *Decoder) decodeVCardFieldsIntoMap(s string) (map[string]string, Schema, string, error) {

	m := make(map[string]string)
	offset := 0

	for line := range strings.Lines(s) {
		trimmed := strings.TrimSpace(line)
		if trimmed == expectedFooter {
			break
		}
		offset += len(line)

		parseErr := parsingErrf("unable to decode line %q. Should have format %q", line, "KEY:VALUE\r\n")

		idx := strings.IndexFunc(trimmed, func(r rune) bool {
			return !unicode.IsLetter(r)
		})
		if idx == -1 {
			return m, Schema{}, s, parseErr
		}

		key := trimmed[:idx]
		value := trimmed[idx:]

		if key == "" || value == "" {
			return m, Schema{}, s, parseErr
		}
		m[key] = value
	}

	s = s[offset:]

	ver, found := m["VERSION"]
	if !found {
		return m, Schema{}, s, parsingErrf("field %q was not found", "VERSION")
	}
	ver = ver[1:]

	schema, found := d.schemas[ver]
	if !found {
		return m, Schema{}, s, parsingErrf("schema for version %q was not provided to Decoder", ver)
	}

	for req := range schema.requiredFields {
		_, found := m[req]
		if !found {
			return m, schema, s, parsingErrf("document does not contain a field %q required by the schema", req)
		}
	}

	return m, schema, s, nil
}

func (d *Decoder) decodeSlice(s string, v reflect.Value) (string, error) {
	panic("TODO: decodeSlice")
}

func (d *Decoder) decodeArray(s string, v reflect.Value) (string, error) {
	panic("TODO: decodeArray")
}

const expectedHeader = "BEGIN:VCARD"

func (d *Decoder) decodeRecordHeader(s string) (string, error) {
	if s == "" {
		return s, parsingErrf("%w", io.ErrUnexpectedEOF)
	}

	lineLen := 0
	for line := range strings.Lines(s) {
		if strings.TrimSpace(line) != expectedHeader {
			return s, parsingErrf("expected %q but found %q", expectedHeader, line)
		}
		lineLen = len(line)
		break
	}

	return s[lineLen:], nil
}

const expectedFooter = "END:VCARD"

func (d *Decoder) decodeRecordFooter(s string) (string, error) {
	if s == "" {
		return s, parsingErrf("%w", io.ErrUnexpectedEOF)
	}

	lineLen := 0
	for line := range strings.Lines(s) {
		if strings.TrimSpace(line) != expectedFooter {
			return s, parsingErrf("expected %q but found %q", expectedFooter, line)
		}
		lineLen = len(line)
		break
	}

	return s[lineLen:], nil
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
