package vcardgo

import (
	"bytes"
	"fmt"
	"io"
)

// Deserializes a vCard document into a Go value.
//
// Short-hand for Decoder.Decode() with default set of schemas.
func Unmarshal(data []byte, v any) error {
	r := bytes.NewReader(data)
	dec := NewDecoder(r, DefaultSchemas)
	return dec.Decode(v)
}

type Decoder struct {
	r       io.Reader
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

func (d *Decoder) Decode(v any) error {
	panic("TODO: Decode")
}
