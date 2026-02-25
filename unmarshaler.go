package vcard

import (
	"bytes"
	"errors"
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
	}
	return s, vCardErrf("unable to decode into an unsupported type %s. Use pointer to a struct, map or a slice", v.Type())
}

func (d *Decoder) decodeMap(data string, ma reflect.Value) (string, error) {
	if ma.IsNil() {
		return data, vCardErrf("decoding is only possible into not-nil map")
	}

	s, err := d.decodeRecordHeader(data)
	if err != nil {
		return data, err
	}
	m, schema, s, err := d.decodeVCardFieldsIntoMap(s)
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
		return s, leftTokensErrf("after successfully decoding a map")
	}

	return s, nil
}

func (d *Decoder) fillMap(ma reflect.Value, m map[string]string, schema Schema) error {

	key := ma.Type().Key()
	if key.Kind() != reflect.String {
		return vCardErrf("unable to decode into a map where key has unsupported type %s. Use string instead", key)
	}

	elemType := ma.Type().Elem()

	switch elemType.Kind() {
	case reflect.String:
		newMap := make(map[string]string, len(schema.fields))

		for req := range schema.fields {
			v, found := m[req]
			if !found {
				continue
			}
			if !d.smartStrings {
				newMap[req] = v
			} else {
				if v[0] == ':' {
					newMap[req] = v[1:]
				} else {
					newMap[req] = v
				}
			}
		}
		ma.Set(reflect.ValueOf(newMap))

	case reflect.Struct:

		if elemType.Implements(reflect.TypeFor[VCardFieldUnmarshaler]()) {
			for field := range schema.fields {
				v, found := m[field]
				if !found {
					continue
				}

				value := reflect.Zero(elemType)
				i := value.Interface().(VCardFieldUnmarshaler)

				err := i.UnmarshalVCardField([]byte(v))
				if err != nil {
					return vCardErrf("error while unmarshaling a value for a key %q: %w", field, err)
				}

				ma.SetMapIndex(reflect.ValueOf(field), value)
			}
		} else if reflect.PointerTo(elemType).Implements(reflect.TypeFor[VCardFieldUnmarshaler]()) {
			for field := range schema.fields {
				v, found := m[field]
				if !found {
					continue
				}

				valuePtr := reflect.New(elemType)
				i := valuePtr.Interface().(VCardFieldUnmarshaler)

				err := i.UnmarshalVCardField([]byte(v))
				if err != nil {
					return vCardErrf("error while unmarshaling a value for a key %q: %w", field, err)
				}

				ma.SetMapIndex(reflect.ValueOf(field), valuePtr.Elem())
			}
		} else {
			return vCardErrf("unable to decode into a map where value has type %s that does not implement VCardFieldUnmarshaler", elemType)
		}

	case reflect.Pointer:
		for field := range schema.fields {
			v, found := m[field]
			if !found {
				continue
			}
			deref := elemType.Elem()

			if deref.Kind() != reflect.Struct {
				return vCardErrf("map value has unsupported pointer type %s. Only struct pointers are allowed", elemType)
			}

			valuePtr := reflect.New(deref)
			i, ok := valuePtr.Interface().(VCardFieldUnmarshaler)
			if !ok {
				return vCardErrf("unable to decode a value for a map key %q because it has type %s which does not implement VCardFieldUnmarshaler", field, elemType)
			}

			err := i.UnmarshalVCardField([]byte(v))
			if err != nil {
				return vCardErrf("error while unmarshaling a value for a key %q: %w", field, err)
			}

			ma.SetMapIndex(reflect.ValueOf(field), valuePtr)
		}

	case reflect.Interface:
		return vCardErrf("unable to decode into a map where value has interface type %s. Use a string or specific struct that implements VCardFieldUnmarshaler instead of interface", elemType)

	default:
		return vCardErrf("unable to decode into a map where value has unsupported type %s. Use string or struct that implements VCardFieldUnmarshaler", elemType)
	}

	return nil
}

func (d *Decoder) decodeStruct(data string, struc reflect.Value) (string, error) {

	s, err := d.decodeRecordHeader(data)
	if err != nil {
		return data, err
	}
	m, schema, s, err := d.decodeVCardFieldsIntoMap(s)
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
		fieldDesc := struc.Type().Field(i)
		fieldValue := struc.Field(i)

		vCardName := fieldDesc.Name

		tag := fieldDesc.Tag.Get("vCard")
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
			return vCardErrf("unable to set a field %q of struct %s. This field is private or unaddressable", fieldDesc.Name, struc.Type())
		}
		taggedMsg := ""
		if tag != "" {
			taggedMsg = fmt.Sprintf("tagged `vCard:\"%s\"` ", tag)
		}

		switch fieldDesc.Type.Kind() {
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
		case reflect.Struct:

			if fieldDesc.Type.Implements(reflect.TypeFor[VCardFieldUnmarshaler]()) {
				for field := range schema.fields {
					v, found := m[field]
					if !found || field != fieldDesc.Name {
						continue
					}

					value := reflect.Zero(fieldDesc.Type)
					i := value.Interface().(VCardFieldUnmarshaler)

					err := i.UnmarshalVCardField([]byte(v))
					if err != nil {
						return vCardErrf("error during unmarshaling field %q %sof struct %s: %w", fieldDesc.Name, taggedMsg, struc.Type(), err)
					}

					fieldValue.Set(value)
				}
			} else if reflect.PointerTo(fieldDesc.Type).Implements(reflect.TypeFor[VCardFieldUnmarshaler]()) {

				for field := range schema.fields {
					v, found := m[field]
					if !found || field != fieldDesc.Name {
						continue
					}

					valuePtr := reflect.New(fieldDesc.Type)
					i := valuePtr.Interface().(VCardFieldUnmarshaler)

					err := i.UnmarshalVCardField([]byte(v))
					if err != nil {
						return vCardErrf("error during unmarshaling field %q %sof struct %s: %w", fieldDesc.Name, taggedMsg, struc.Type(), err)
					}

					fieldValue.Set(valuePtr.Elem())
				}
			} else {
				return vCardErrf("field %q %sof type %s has unsupported type %s. Use string or struct that implements VCardFieldUnmarshaler", fieldDesc.Name, taggedMsg, struc.Type(), fieldDesc.Type)
			}

		case reflect.Pointer:
			if fieldDesc.Type.Implements(reflect.TypeFor[VCardFieldUnmarshaler]()) {

				for field := range schema.fields {
					v, found := m[field]
					if !found || field != fieldDesc.Name {
						continue
					}

					valuePtr := reflect.New(fieldDesc.Type.Elem())
					i := valuePtr.Interface().(VCardFieldUnmarshaler)

					err := i.UnmarshalVCardField([]byte(v))
					if err != nil {
						return vCardErrf("error during unmarshaling field %q %sof struct %s: %w", fieldDesc.Name, taggedMsg, struc.Type(), err)
					}

					fieldValue.Set(valuePtr)
				}
			} else {
				return vCardErrf("field %q %sof type %s has unsupported pointer type %s. Use string or struct that implements VCardFieldUnmarshaler", fieldDesc.Name, taggedMsg, struc.Type(), fieldDesc.Type)
			}

		case reflect.Interface:
			return vCardErrf("field %q %sof type %s has interface type %s. Use string or specific struct that implements VCardFieldUnmarshaler", fieldDesc.Name, taggedMsg, struc.Type(), fieldDesc.Type)

		default:
			return vCardErrf("field %q %sof type %s has unsupported type %s. Use string or struct that implements VCardFieldUnmarshaler", fieldDesc.Name, taggedMsg, struc.Type(), fieldDesc.Type)
		}
	}

	return nil
}

func (d *Decoder) decodeVCardFieldsIntoMap(s string) (map[string]string, Schema, string, error) {

	m := make(map[string]string)
	offset := 0

	lastKey := ""
	for line := range strings.Lines(s) {
		trimmed := strings.TrimSpace(line)
		if trimmed == expectedFooter {
			break
		}
		offset += len(line)

		if trimmed == "" {
			continue
		}

		isContinuation := false
		if len(line) >= 1 && line[0] == ' ' {
			isContinuation = true
		}

		if isContinuation {
			lastValue := m[lastKey]
			lastValue = lastValue + trimmed
			m[lastKey] = lastValue

			continue
		}

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

		lastKey = key
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

func (d *Decoder) decodeSlice(data string, slice reflect.Value) (string, error) {

	elemDesc := slice.Type().Elem()

	switch elemDesc.Kind() {
	case reflect.Struct:

		reflectStructsSlice := make([]reflect.Value, 0)

		s := data
		for {
			ptr := reflect.New(elemDesc)
			struc := ptr.Elem()

			var err error
			s, err = d.decodeStruct(s, struc)

			reflectStructsSlice = append(reflectStructsSlice, struc)

			// No tokens left after decoding means success
			if err == nil {
				break
			}

			if !errors.Is(err, ErrLeftoverTokens) {
				return s, err
			}
		}

		outSlice := reflect.MakeSlice(slice.Type(), len(reflectStructsSlice), len(reflectStructsSlice))

		for i, reflectStruct := range reflectStructsSlice {
			outSlice.Index(i).Set(reflectStruct)
		}

		slice.Set(outSlice)

		return s, nil

	case reflect.Map:

		reflectMapsSlice := make([]reflect.Value, 0)

		s := data
		for {
			ptr := reflect.New(elemDesc)
			m := ptr.Elem()
			m.Set(reflect.MakeMap(elemDesc))

			var err error
			s, err = d.decodeMap(s, m)

			reflectMapsSlice = append(reflectMapsSlice, m)

			// No tokens left after decoding means success
			if err == nil {
				break
			}

			if !errors.Is(err, ErrLeftoverTokens) {
				return s, err
			}
		}

		outSlice := reflect.MakeSlice(slice.Type(), len(reflectMapsSlice), len(reflectMapsSlice))

		for i, reflectMap := range reflectMapsSlice {
			outSlice.Index(i).Set(reflectMap)
		}

		slice.Set(outSlice)

		return s, nil

	default:
		return data, vCardErrf("unable to decode into a slice where element has type %s. Use slice of maps or structs that implement VCardFieldUnmarshaler", elemDesc)
	}
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
