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

type VersionUser struct {
	N    string
	FN   string
	NAME string
	VERSION string
}

func TestDecStructStringFieldsWithVersion(t *testing.T) {

	s := VersionUser{}

	text := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	_ = Unmarshal([]byte(text), &s)

	exp := VersionUser{
		N:    "Alex",
		FN:   "Alex FullName",
		NAME: "Alex Name Hello",
		VERSION: "4.0",
	}

	assertEq(t, s, exp)
}

func TestDecMapStringString(t *testing.T) {

	s := make(map[string]string)

	text := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	_ = Unmarshal([]byte(text), &s)

	exp := map[string]string {
		"VERSION": ":4.0",
		"N": ":Alex",
		"FN": ":Alex FullName",
		"NAME": ":Alex Name Hello",
	}

	assertMapsEq(t, s, exp)
}