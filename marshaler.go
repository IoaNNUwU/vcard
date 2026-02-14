package vcardgo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Serializes a Go value as a vCard document.
//
// Short-hand for Encoder.Encode() with default vCard 4.0 schema.
//
// fields of v have to either match the name and the type from the schema or implement
// Marshaler for custom encoding logic.
func Marshal(v any) ([]byte, error) {
	return MarshalSchema(v, DefaultSchemaV4)
}

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

	smartStrings bool

	// TODO: Cache prepared schema between EncodeSchema() calls
	// TODO: Cache type info between encode() calls
}

// Creates new Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		smartStrings: true,
		w:            w,
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
// You can overcome this limitation by using custom fields implementing VCardMarshaler instead of string.
func (e *Encoder) SetSmartStrings(smartStrings bool) {
	e.smartStrings = smartStrings
}

// Writes a vCard representation of v to the stream using default vCard 4.0 schema.
//
// fields of v have to either match the name and the type from the schema or implement
// Marshaler for custom encoding logic.
func (e *Encoder) Encode(v any) error {
	return e.EncodeSchema(v, DefaultSchemaV4)
}

// Writes a vCard representation of v to the stream using provided Schema.
//
// v has to be a struct, slice of structs, map or slice of maps. Map key has to be a string.
// Map value has to either implement [VCardFieldMarshaler] or be one of the supported types,
// e.g. a string.
func (e *Encoder) EncodeSchema(v any, schema Schema) error {
	if v == nil {
		return errors.New("vCard: cannot encode a nil interface")
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
		return fmt.Errorf("vCard: cannot write: %w", err)
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
	return b, errors.New("vCard: unable to encode %s type. Use struct, map or a slice")
}

func (e *Encoder) encodeMap(b []byte, ma reflect.Value, ctx encoderCtx) ([]byte, error) {
	keyKind := ma.Type().Key().Kind()
	if keyKind != reflect.String {
		return []byte{}, fmt.Errorf("vCard: type %s is not supported as a map key. Use string instead", keyKind)
	}

	for req := range ctx.schema.requiredFields {
		if !ma.MapIndex(reflect.ValueOf(req)).IsValid() {
			return b, fmt.Errorf("vCard: map does not contain field `%s` required by the schema", req)
		}
	}
	buf := e.encodeRecordHeader([]byte{}, ctx)

	i := ma.MapRange()

	// In case of an empty map lets write BEGIN:VCARD, VERSION:.. and END:VCARD to simplify debugging
	// This is only possible for user-defined schema with no required fields
	if !i.Next() {
		buf = e.encodeRecordFooter(buf)
		return append(b, buf...), nil
	}
	// It's better to inspect kind of the first element single time at the start
	valueKind := i.Value().Kind()

	switch valueKind {
	case reflect.String:
		m := ma.Interface().(map[string]string)

		if !e.smartStrings {
			for k, v := range m {
				buf = append(buf, fmt.Sprintf("%s%s\n", k, v)...)
			}
		} else {
			for k, v := range m {
				if strings.Contains(v, ":") {
					buf = append(buf, fmt.Sprintf("%s%s\n", k, v)...)
				} else {
					buf = append(buf, fmt.Sprintf("%s:%s\n", k, v)...)
				}
			}
		}
	case reflect.Struct:
		if i.Value().Type().Implements(vCardFieldMarshaler) {
			iter := ma.MapRange()
			for iter.Next() {
				k := iter.Key().String()
				v := iter.Value().Interface().(VCardFieldMarshaler)

				field, err := v.MarshalVCardField()
				if err != nil {
					return b, fmt.Errorf("vCard: error during marshaling value for a key `%s`: %w", k, err)
				}
				buf = append(buf, fmt.Sprintf("%s%s\n", k, field)...)
			}
		} else {
			return b, fmt.Errorf("vCard: map value is a struct of type %s which does not implement VCardFieldMarshaler", i.Value().Type())
		}
	case reflect.Interface:
		iter := ma.MapRange()
		for iter.Next() {
			k := iter.Key().String()
			v, ok := iter.Value().Interface().(VCardFieldMarshaler)

			if !ok {
				return b, fmt.Errorf("vCard: map value for a key `%s` is a struct of type %s which does not implement VCardFieldMarshaler", k, iter.Value().Type())
			}

			field, err := v.MarshalVCardField()
			if err != nil {
				return b, fmt.Errorf("vCard: error during marshaling value for a key `%s`: %w", k, err)
			}
			buf = append(buf, fmt.Sprintf("%s%s\n", k, field)...)
		}
	default:
		return b, fmt.Errorf("vCard: type %s is not supported as a map value. Use string or a struct that implements VCardFieldMarshaler", i.Value().Type())
	}
	buf = e.encodeRecordFooter(buf)

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
			return b, fmt.Errorf("vCard: struct %v does not contain field `%s` or field tagged `vCard:\"%s\"` required by the schema", struc.Type(), req, req)
		}
	}
	buf := e.encodeRecordHeader([]byte{}, ctx)

	// In case of an empty struct lets write BEGIN:VCARD, VERSION:.. and END:VCARD to simplify debugging
	// This is only possible for user-defined schema with no required fields
	if struc.NumField() == 0 {
		buf = e.encodeRecordFooter(buf)
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

		_, exists := ctx.schema.fields[vCardName]
		if !exists {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			s := field.String()
			if !e.smartStrings {
				buf = append(buf, fmt.Sprintf("%s%s\n", vCardName, s)...)
			} else {
				if strings.Contains(s, ":") {
					buf = append(buf, fmt.Sprintf("%s%s\n", vCardName, s)...)
				} else {
					buf = append(buf, fmt.Sprintf("%s:%s\n", vCardName, s)...)
				}
			}
		case reflect.Struct, reflect.Interface:
			v, ok := field.Interface().(VCardFieldMarshaler)

			if !ok {
				return b, fmt.Errorf("vCard: field `%s` %sof a struct %s has type %s which does not implement VCardFieldMarshaler", fieldDesc.Name, taggedMsg, struc.Type(), field.Type())
			}

			fieldBytes, err := v.MarshalVCardField()
			if err != nil {
				return b, fmt.Errorf("vCard: error during marshaling field `%s` %sof struct %s: %w", fieldDesc.Name, taggedMsg, struc.Type(), err)
			}
			buf = append(buf, fmt.Sprintf("%s%s\n", vCardName, fieldBytes)...)

		default:
			return b, fmt.Errorf("vCard: field `%s` %sof a struct %s has type %s which is not supported. Use string or a struct that implements VCardFieldMarshaler", fieldDesc.Name, taggedMsg, struc.Type(), field.Type())
		}
	}

	buf = e.encodeRecordFooter(buf)

	return append(b, buf...), nil
}

func (e *Encoder) encodeRecordHeader(b []byte, ctx encoderCtx) []byte {
	return append(b, fmt.Sprintf("BEGIN:VCARD\nVERSION:%s\n", ctx.schema.version)...)
}

func (e *Encoder) encodeRecordFooter(b []byte) []byte {
	return append(b, "END:VCARD\n"...)
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
				return b, fmt.Errorf("vCard: error during marshaling slice member idx=%v: %w", i, err)
			}
		}
	case reflect.Struct:
		for i := range slice.Len() {
			elem := slice.Index(i)
			var err error
			buf, err = e.encodeStruct(buf, elem, ctx)
			if err != nil {
				return b, fmt.Errorf("vCard: error during marshaling slice member idx=%v: %w", i, err)
			}
		}
	case reflect.Interface:
		for i := range slice.Len() {
			elem := slice.Index(i)
			var err error
			buf, err = e.encode(buf, elem, ctx)
			if err != nil {
				return b, fmt.Errorf("vCard: error during marshaling slice member idx=%v: %w", i, err)
			}
		}
	default:
		return b, fmt.Errorf("vCard: unable to encode slice of type %s. Use slice of structs or maps", elemKind)
	}
	return append(b, buf...), nil
}

type encoderCtx struct {
	schema Schema
}

var vCardFieldMarshaler = reflect.TypeFor[VCardFieldMarshaler]()

// Implemented by fields that need custom Marshaling logic.
//
// Note that this interface defines a way to marshal single field.
// For example TEL field in your schema has custom type Tel:
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
