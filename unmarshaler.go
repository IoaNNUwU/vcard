package vcardgo

import (
	"bytes"
	"io"
)

// Deserializes a vCard document into a Go value.
//
// Short-hand for Decoder.Decode() with default schema.
// Each field has to be a string in this case.
func Unmarshal(data []byte, v any) error {
	r := bytes.NewReader(data)
	dec := NewDecoder(r, defaultSchemas)
	return dec.Decode(v)
}

type Decoder struct {
	r       io.Reader
	schemas map[string]Schema
}

var defaultSchemas = []Schema{
	DefaultSchemaV4,
	DefaultSchemaV3,
	DefaultSchemaV2_1,
}

func NewDecoder(r io.Reader, schemas []Schema) *Decoder {
	m := make(map[string]Schema)

	for _, s := range schemas {
		m[s.version] = s
	}
	return &Decoder{
		r:       r,
		schemas: m,
	}
}

func (d *Decoder) Decode(v any) error {
	panic("TODO: Decode")
}
