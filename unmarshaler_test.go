package vcard

import "testing"

func TestDecEmptyStruct(t *testing.T) {

	e := &Empty{}

	ser := `BEGIN:VCARD
VERSION:4.0
END:VCARD
`

	err := UnmarshalSchema([]byte(ser), e, []Schema{EmptySchema})

	assertEq(t, err, nil)
}

func TestDecStructStringFields(t *testing.T) {

	s := StringUser{}

	text := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	_ = Unmarshal([]byte(text), &s)

	exp := StringUser{
		N:    "Alex",
		FN:   "Alex FullName",
		NAME: "Alex Name Hello",
	}

	assertEq(t, s, exp)
}
