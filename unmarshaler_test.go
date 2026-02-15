package vcard

import "testing"

func TestUnmarshalEmptyStruct(t *testing.T) {

	e := &Empty{}

	ser := `BEGIN:VCARD
VERSION:4.0
END:VCARD
`

	err := Unmarshal([]byte(ser), e)

	AssertErr(t, err)
	AssertStringContains(t, err.Error(), "Skibidi")
}
