package vcard

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Serializes a Go value as a vCard document using default vCard 4.0 schema.
//
// v has to be a map, struct or a slice.
func Marshal(v any) ([]byte, error) {
	return MarshalSchema(v, SchemaV4)
}

// Serializes a Go value as a vCard document using provided [Schema].
//
// v has to be a map, struct or a slice.
func MarshalSchema(v any, schema Schema) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	err := enc.EncodeSchema(v, schema)
	if err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

// Writes a vCard document to an output stream.
type Encoder struct {
	w io.Writer

	smartStrings    bool
	newlineSequence string

	// TODO: Cache prepared schema between EncodeSchema() calls
	// TODO: Cache type info between encode() calls
}

// Creates new Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:               w,
		smartStrings:    true,
		newlineSequence: "\r\n",
	}
}

// Toggles smart string encoding. Enabled by default.
//
// In smart mode, encoder checks at runtime if string contains `:` (KEY:VALUE separator) and adds
// it if neccesary. This is useful because some fields have more complex format e.g.:
//
// For string "N:;Alex;;;" k="N", v=";Alex;;;" - `:` won't be a part of an encoded value.
//
// For string "TEL;TYPE=CELL:555" k="TEL", v=";TYPE=CELL:555" - `:` will be in the middle
// of an encoded value.
//
// Disabling smart strings encoding will increase performance, but you have to ensure your
// strings have proper puctuation in them e.g. you will have to deal with ":Name" instead of "Name".
//
// You can overcome this limitation by using custom fields implementing [VCardFieldMarshaler] instead of string.
func (e *Encoder) SetSmartStrings(smartStrings bool) *Encoder {
	e.smartStrings = smartStrings
	return e
}

// Sets newline sequence. Defaults to "\r\n" as per vCard RFC https://datatracker.ietf.org/doc/html/rfc6350#section-3.2
func (e *Encoder) SetNewlineSequence(seq string) *Encoder {
	e.newlineSequence = seq
	return e
}

// Writes a vCard representation of v to the stream using default vCard 4.0 schema.
//
// fields of v have to either match the name and the type from the schema or implement
// Marshaler for custom encoding logic.
func (e *Encoder) Encode(v any) error {
	return e.EncodeSchema(v, SchemaV4)
}

// Writes a vCard representation of v to the stream using provided Schema.
//
// v has to be a struct, slice of structs, map or slice of maps. Map key has to be a string.
// Map value has to either implement [VCardFieldMarshaler] or be one of the supported types,
// e.g. a string.
func (e *Encoder) EncodeSchema(v any, schema Schema) error {
	if v == nil {
		return vCardErrf("cannot encode a nil interface")
	}
	// Intermidiate buffer makes sure there was no errors before writing to io.Writer
	b := []byte{}

	// TODO: Cache prepared schema between EncodeSchema() calls
	ctx := encoderCtx{schema: schema}

	b, err := e.encode(b, reflect.ValueOf(v), ctx)
	if err != nil {
		return err
	}
	_, err = e.w.Write(b)
	if err != nil {
		return vCardErrf("cannot write: %w", err)
	}
	return nil
}

func (e *Encoder) encode(b []byte, v reflect.Value, ctx encoderCtx) ([]byte, error) {
	switch v.Kind() {
	case reflect.Map:
		return e.encodeMap(b, v, ctx)
	case reflect.Struct:
		return e.encodeStruct(b, v, ctx)
	case reflect.Array, reflect.Slice:
		return e.encodeSlice(b, v, ctx)
	}
	return b, vCardErrf("unable to encode %s type. Use struct, map or a slice", v.Type())
}

func (e *Encoder) encodeMap(b []byte, ma reflect.Value, ctx encoderCtx) ([]byte, error) {
	keyKind := ma.Type().Key().Kind()
	if keyKind != reflect.String {
		return []byte{}, vCardErrf("type %s is not supported as a map key. Use string instead", keyKind)
	}
	for req := range ctx.schema.requiredFields {
		if !ma.MapIndex(reflect.ValueOf(req)).IsValid() {
			return b, vCardErrf("map does not contain field %q required by the schema", req)
		}
	}
	buf := e.encodeRecordHeader([]byte{}, ctx)

	i := ma.MapRange()

	// In case of an empty map lets write BEGIN:VCARD, VERSION:.. and END:VCARD to simplify debugging
	// This is only possible for user-defined schema with no required fields
	if !i.Next() {
		buf = e.encodeRecordFooter(buf, ctx)
		return append(b, buf...), nil
	}
	// It's better to inspect kind of the first element single time at the start
	valueKind := i.Value().Kind()

	switch valueKind {
	case reflect.String:
		m := ma.Interface().(map[string]string)

		if !e.smartStrings {
			for k, v := range m {
				_, found := ctx.schema.fields[k]
				if !found {
					continue
				}
				buf = append(buf, fmt.Sprintf("%s%s%s", k, v, e.newlineSequence)...)
			}
		} else {
			for k, v := range m {
				_, found := ctx.schema.fields[k]
				if !found {
					continue
				}
				if strings.Contains(v, ":") {
					buf = append(buf, fmt.Sprintf("%s%s%s", k, v, e.newlineSequence)...)
				} else {
					buf = append(buf, fmt.Sprintf("%s:%s%s", k, v, e.newlineSequence)...)
				}
			}
		}
	case reflect.Struct:
		if i.Value().Type().Implements(reflect.TypeFor[VCardFieldMarshaler]()) {
			iter := ma.MapRange()
			for iter.Next() {
				k := iter.Key().String()

				_, found := ctx.schema.fields[k]
				if !found {
					continue
				}

				v := iter.Value().Interface().(VCardFieldMarshaler)

				field, err := v.MarshalVCardField()
				if err != nil {
					return b, vCardErrf("error during marshaling value for a key %q: %w", k, err)
				}
				buf = append(buf, fmt.Sprintf("%s%s%s", k, field, e.newlineSequence)...)
			}
		} else {
			return b, vCardErrf("map value is a struct of type %s which does not implement VCardFieldMarshaler", i.Value().Type())
		}
	case reflect.Interface:
		iter := ma.MapRange()
		for iter.Next() {
			k := iter.Key().String()
			v, ok := iter.Value().Interface().(VCardFieldMarshaler)

			if !ok {
				return b, vCardErrf("map value for a key %q is a struct of type %s which does not implement VCardFieldMarshaler", k, iter.Value().Type())
			}
			field, err := v.MarshalVCardField()
			if err != nil {
				return b, vCardErrf("error during marshaling value for a key %q: %w", k, err)
			}
			buf = append(buf, fmt.Sprintf("%s%s%s", k, field, e.newlineSequence)...)
		}
	default:
		return b, vCardErrf("type %s is not supported as a map value. Use string or a struct that implements VCardFieldMarshaler", i.Value().Type())
	}
	buf = e.encodeRecordFooter(buf, ctx)

	return append(b, buf...), nil
}

func (e *Encoder) encodeStruct(b []byte, struc reflect.Value, ctx encoderCtx) ([]byte, error) {

	// TODO: Cache struct fields lookup
	for req := range ctx.schema.requiredFields {

		structField, _ := struc.Type().FieldByName(req)
		fieldName := structField.Name

		// Check for another field tagged `vCard:"N"`
		// which has a priority above field `N`
		for i := range struc.NumField() {
			otherStructField := struc.Type().Field(i)
			tag := otherStructField.Tag.Get("vCard")
			if tag == req {
				fieldName = otherStructField.Name
			}
		}

		if fieldName == "" {
			return b, vCardErrf("struct %v does not contain field %q or field tagged `vCard:\"%s\"` required by the schema", struc.Type(), req, req)
		}
	}
	buf := e.encodeRecordHeader([]byte{}, ctx)

	// In case of an empty struct lets write BEGIN:VCARD, VERSION:.. and END:VCARD to simplify debugging
	// This is only possible for user-defined schema with no required fields
	if struc.NumField() == 0 {
		buf = e.encodeRecordFooter(buf, ctx)
		return append(b, buf...), nil
	}

	for i := range struc.NumField() {

		field := struc.Field(i)
		fieldDesc := struc.Type().Field(i)

		vCardName := fieldDesc.Name

		tag := fieldDesc.Tag.Get("vCard")
		taggedMsg := ""
		if tag != "" {
			vCardName = tag
			taggedMsg = fmt.Sprintf("tagged `vCard:\"%s\"` ", vCardName)
		}

		_, found := ctx.schema.fields[vCardName]
		if !found {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			s := field.String()
			if !e.smartStrings {
				buf = append(buf, fmt.Sprintf("%s%s%s", vCardName, s, e.newlineSequence)...)
			} else {
				if strings.Contains(s, ":") {
					buf = append(buf, fmt.Sprintf("%s%s%s", vCardName, s, e.newlineSequence)...)
				} else {
					buf = append(buf, fmt.Sprintf("%s:%s%s", vCardName, s, e.newlineSequence)...)
				}
			}
		case reflect.Struct, reflect.Interface:
			v, ok := field.Interface().(VCardFieldMarshaler)

			if !ok {
				return b, vCardErrf("field %q %sof a struct %s has type %s which does not implement VCardFieldMarshaler", fieldDesc.Name, taggedMsg, struc.Type(), field.Type())
			}

			fieldBytes, err := v.MarshalVCardField()
			if err != nil {
				return b, vCardErrf("error during marshaling field %q %sof struct %s: %w", fieldDesc.Name, taggedMsg, struc.Type(), err)
			}
			buf = append(buf, fmt.Sprintf("%s%s%s", vCardName, fieldBytes, e.newlineSequence)...)

		default:
			return b, vCardErrf("field %q %sof a struct %s has unsupported type %s. Use string or a struct that implements VCardFieldMarshaler", fieldDesc.Name, taggedMsg, struc.Type(), field.Type())
		}
	}

	buf = e.encodeRecordFooter(buf, ctx)

	return append(b, buf...), nil
}

func (e *Encoder) encodeRecordHeader(b []byte, ctx encoderCtx) []byte {
	return append(b, fmt.Sprintf("BEGIN:VCARD%sVERSION:%s%s", e.newlineSequence, ctx.schema.version, e.newlineSequence)...)
}

func (e *Encoder) encodeRecordFooter(b []byte, _ encoderCtx) []byte {
	return append(b, fmt.Sprintf("END:VCARD%s", e.newlineSequence)...)
}

func (e *Encoder) encodeSlice(b []byte, slice reflect.Value, ctx encoderCtx) ([]byte, error) {
	if slice.Len() == 0 {
		return b, nil
	}
	// Intermidiate buffer makes sure there was no errors before writing bytes
	buf := []byte{}

	elemKind := slice.Index(0).Kind()

	switch elemKind {
	case reflect.Map:
		for i := range slice.Len() {
			elem := slice.Index(i)
			var err error
			buf, err = e.encodeMap(buf, elem, ctx)
			if err != nil {
				return b, vCardErrf("error during marshaling slice member idx=%v: %w", i, err)
			}
		}
	case reflect.Struct:
		for i := range slice.Len() {
			elem := slice.Index(i)
			var err error
			buf, err = e.encodeStruct(buf, elem, ctx)
			if err != nil {
				return b, vCardErrf("error during marshaling slice member idx=%v: %w", i, err)
			}
		}
	case reflect.Interface:
		for i := range slice.Len() {
			elem := slice.Index(i)
			var err error
			buf, err = e.encode(buf, elem, ctx)
			if err != nil {
				return b, vCardErrf("error during marshaling slice member idx=%v: %w", i, err)
			}
		}
	default:
		return b, vCardErrf("unable to encode slice of type %s. Use slice of structs or maps", elemKind)
	}
	return append(b, buf...), nil
}

type encoderCtx struct {
	schema Schema
}

// Implemented by fields that need custom Marshaling logic.
//
// Note that this interface defines a way to marshal a value of single field.
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
//		func (t Tel) MarshalVCardField() ([]byte, error) {
//			// final result is TEL;TYPE=CELL:(123) 555-5832
//			return fmt.Sprintf(";TYPE=%s:%s", typ, tel), nil
//		}
type VCardFieldMarshaler interface {
	MarshalVCardField() ([]byte, error)
}
