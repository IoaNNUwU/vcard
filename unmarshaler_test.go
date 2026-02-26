package vcard

import "testing"

func TestDecIntoNonPointer(t *testing.T) {

	s := `BEGIN:VCARD
END:VCARD
`
	i := 10

	err := Unmarshal([]byte(s), i)

	assertErrIs(t, err, ErrVCard, "decoding is only possible into a pointer, not int")
}

func TestDecIntoNotSupportedPointer(t *testing.T) {

	s := `BEGIN:VCARD
END:VCARD
`
	i := 10

	err := Unmarshal([]byte(s), &i)

	assertErrIs(t, err, ErrVCard, "unable to decode into an unsupported type int")
}

// Map Marshaling Tests

func TestVCardMissingRequiredKey(t *testing.T) {

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
` // missing field TEL

	m := make(map[string]string)

	err := UnmarshalSchema([]byte(s), &m, []Schema{SchemaFor[telRequiredSchema]("4.0")})

	assertErrIs(t, err, ErrVCard, "does not contain a field \"TEL\"")
}

func TestDecEmptyMap(t *testing.T) {

	m := make(map[string]string)

	s := `BEGIN:VCARD
VERSION:4.0
END:VCARD
`
	_ = UnmarshalSchema([]byte(s), &m, []Schema{EmptySchema})

	exp := map[string]string{}

	assertMapsEq(t, m, exp)
}

func TestDecMapStringString(t *testing.T) {

	m := make(map[string]string)

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	_ = Unmarshal([]byte(s), &m)

	exp := map[string]string{
		"VERSION": "4.0",
		"N":       "Alex",
		"FN":      "Alex FullName",
		"NAME":    "Alex Name Hello",
	}

	assertMapsEq(t, m, exp)
}

type schemaWithoutVersion struct {
	N    string `vCard:"required"`
	FN   string
	NAME string
}

func TestDecMapStringStringWithoutVersion(t *testing.T) {

	m := make(map[string]string)

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	_ = UnmarshalSchema([]byte(s), &m, []Schema{SchemaFor[schemaWithoutVersion]("4.0")})

	exp := map[string]string{
		"N":    "Alex",
		"FN":   "Alex FullName",
		"NAME": "Alex Name Hello",
	}

	assertMapsEq(t, m, exp)
}

func TestDecMapSmartString(t *testing.T) {

	m := make(map[string]string)

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
TEL;TYPE=CELL:(555) 333 222
END:VCARD
`
	_ = Unmarshal([]byte(s), &m)

	exp := map[string]string{
		"VERSION": "4.0",
		"N":       "Alex",
		"FN":      "Alex FullName",
		"NAME":    "Alex Name Hello",
		"TEL":     ";TYPE=CELL:(555) 333 222",
	}

	assertMapsEq(t, m, exp)
}

type unmarshalVCardImpl struct {
	s string
}

func (s *unmarshalVCardImpl) UnmarshalVCardField(data []byte) error {
	s.s = string(data)
	return nil
}

func TestDecMapCustomUnmarshalerPointer(t *testing.T) {

	m := make(map[string]*unmarshalVCardImpl)

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	_ = Unmarshal([]byte(s), &m)

	exp := map[string]*unmarshalVCardImpl{
		"VERSION": {":4.0"},
		"N":       {":Alex"},
		"FN":      {":Alex FullName"},
		"NAME":    {":Alex Name Hello"},
	}

	assertMapsEq(t, m, exp)
}

func TestDecMapCustomUnmarshaler(t *testing.T) {

	m := make(map[string]unmarshalVCardImpl)

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	_ = Unmarshal([]byte(s), &m)

	exp := map[string]unmarshalVCardImpl{
		"VERSION": {":4.0"},
		"N":       {":Alex"},
		"FN":      {":Alex FullName"},
		"NAME":    {":Alex Name Hello"},
	}

	assertMapsEq(t, m, exp)
}

func TestDecInterfaceCustomUnmarshaler(t *testing.T) {

	m := make(map[string]VCardFieldUnmarshaler)

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	err := Unmarshal([]byte(s), &m)

	assertErrIs(t, err, ErrVCard, "unable to decode into a map where value has interface type")
}

// Struct Marshaling Tests

func TestDecEmptyStruct(t *testing.T) {

	e := &empty{}

	s := `BEGIN:VCARD
VERSION:4.0
END:VCARD
`
	err := UnmarshalSchema([]byte(s), e, []Schema{EmptySchema})

	assertEq(t, err, nil)
}

func TestDecStructStringFields(t *testing.T) {

	u := stringUser{}

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	_ = Unmarshal([]byte(s), &u)

	exp := stringUser{
		N:    "Alex",
		FN:   "Alex FullName",
		NAME: "Alex Name Hello",
	}

	assertEq(t, u, exp)
}

type versionUser struct {
	N       string
	FN      string
	NAME    string
	VERSION string
}

func TestDecStructStringFieldsWithVersion(t *testing.T) {

	u := versionUser{}

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	_ = Unmarshal([]byte(s), &u)

	exp := versionUser{
		N:       "Alex",
		FN:      "Alex FullName",
		NAME:    "Alex Name Hello",
		VERSION: "4.0",
	}

	assertEq(t, u, exp)
}

type unmarshalerUser struct {
	FN   unmarshalerPtrImpl
	N    unmarshalerPtrImpl
	NAME unmarshalerPtrImpl
}

type unmarshalerPtrImpl struct {
	s string
}

func (u *unmarshalerPtrImpl) UnmarshalVCardField(data []byte) error {
	u.s = string(data)
	return nil
}

func (u *unmarshalerPtrImpl) String() string {
	return u.s
}

func TestDecStructUnmarshalerFields(t *testing.T) {

	u := unmarshalerUser{}

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	_ = Unmarshal([]byte(s), &u)

	exp := unmarshalerUser{
		N:    unmarshalerPtrImpl{":Alex"},
		FN:   unmarshalerPtrImpl{":Alex FullName"},
		NAME: unmarshalerPtrImpl{":Alex Name Hello"},
	}

	assertEq(t, u, exp)
}

type unmarshalerImplNoPtr struct {
	s *string // nolint
}

type unmarshalerUserNoPtr struct {
	FN   unmarshalerImplNoPtr
	N    unmarshalerImplNoPtr
	NAME unmarshalerImplNoPtr
}

func (u unmarshalerImplNoPtr) UnmarshalVCardField(data []byte) error {
	// *u.s = string(data)
	return nil
}

func TestDecStructPtrUnmarshalerFields(t *testing.T) {

	u := unmarshalerUserNoPtr{}

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	_ = Unmarshal([]byte(s), &u)

	exp := unmarshalerUserNoPtr{}

	assertEq(t, u, exp)
}

type unmarshalerUserPtrFields struct {
	FN   *unmarshalerPtrImpl
	N    *unmarshalerPtrImpl
	NAME *unmarshalerPtrImpl
}

func TestDecStructUnmarshalerPtrFields(t *testing.T) {

	u := unmarshalerUserPtrFields{}

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	_ = Unmarshal([]byte(s), &u)

	assertEq(t, u.N.s, ":Alex")
	assertEq(t, u.FN.s, ":Alex FullName")
	assertEq(t, u.NAME.s, ":Alex Name Hello")
}

func TestDecStructHasRenameTags(t *testing.T) {

	u := tagsUser{}

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	_ = Unmarshal([]byte(s), &u)

	exp := tagsUser{
		Name:        "Alex",
		FullName:    "Alex FullName",
		Description: "Alex Name Hello",
	}

	assertEq(t, u, exp)
}

func TestDecStructDoesNotImplementUnmarshaler(t *testing.T) {

	u := unsupportedUser{}

	s := `BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
`
	err := Unmarshal([]byte(s), &u)

	assertErrIs(t, err, ErrVCard, "unsupported type")
}

type customUnmarshalerUser struct {
	N    unmarshalVCardImpl
	FN   unmarshalVCardImpl
	NAME unmarshalVCardImpl
}

func TestUnmarshalSliceOfStructs(t *testing.T) {

	s := []byte(`BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
BEGIN:VCARD
VERSION:4.0
N:John
FN:John FullName
NAME:John Name Hello
END:VCARD
BEGIN:VCARD
VERSION:4.0
N:Jane
FN:Jane FullName
NAME:Jane Name Hello
END:VCARD`)

	users := make([]customUnmarshalerUser, 3)
	_ = Unmarshal(s, &users)

	assertEq(t, users[0].N.s, ":Alex")
	assertEq(t, users[0].FN.s, ":Alex FullName")
	assertEq(t, users[0].NAME.s, ":Alex Name Hello")

	assertEq(t, users[1].N.s, ":John")
	assertEq(t, users[1].FN.s, ":John FullName")
	assertEq(t, users[1].NAME.s, ":John Name Hello")

	assertEq(t, users[2].N.s, ":Jane")
	assertEq(t, users[2].FN.s, ":Jane FullName")
	assertEq(t, users[2].NAME.s, ":Jane Name Hello")
}

func TestUnmarshalSliceOfMaps(t *testing.T) {

	s := []byte(`BEGIN:VCARD
VERSION:4.0
N:Alex
FN:Alex FullName
NAME:Alex Name Hello
END:VCARD
BEGIN:VCARD
VERSION:4.0
N:John
FN:John FullName
NAME:John Name Hello
END:VCARD
BEGIN:VCARD
VERSION:4.0
N:Jane
FN:Jane FullName
NAME:Jane Name Hello
END:VCARD`)

	users := make([]map[string]unmarshalerPtrImpl, 3)
	_ = Unmarshal(s, &users)

	exp := []map[string]unmarshalerPtrImpl{
		{
			"N":       {s: ":Alex"},
			"FN":      {s: ":Alex FullName"},
			"NAME":    {s: ":Alex Name Hello"},
			"VERSION": {s: ":4.0"},
		},
		{
			"N":       {s: ":John"},
			"FN":      {s: ":John FullName"},
			"NAME":    {s: ":John Name Hello"},
			"VERSION": {s: ":4.0"},
		},
		{
			"N":       {s: ":Jane"},
			"FN":      {s: ":Jane FullName"},
			"NAME":    {s: ":Jane Name Hello"},
			"VERSION": {s: ":4.0"},
		},
	}

	assertMapsEq(t, users[0], exp[0])
	assertMapsEq(t, users[1], exp[1])
	assertMapsEq(t, users[2], exp[2])
}
