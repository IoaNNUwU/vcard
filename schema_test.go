package vcardgo

import (
	"reflect"
	"testing"
)

type TestImplementation struct {
	N    string
	NAME string `vCard:"required"`
	FN   string `vCard:"required"`
}

func TestSchemaToMap(t *testing.T) {

	schema := NewSchema[TestImplementation]("3.0")

	m := schema.Prepare().hmap

	exp := map[string]SchemaField{
		"N":    {typ: reflect.TypeFor[string](), optional: true},
		"NAME": {typ: reflect.TypeFor[string](), optional: false},
		"FN":   {typ: reflect.TypeFor[string](), optional: false},
	}
	AssertMapsEq(t, m, exp)
}

func TestSchemaToRequiredSlice(t *testing.T) {

	schema := NewSchema[TestImplementation]("3.0")

	s := schema.Prepare().requiredFields

	exp := []string{"NAME", "FN"}

	AssertSlicesEq(t, s, exp)
}

func TestEmptySchemaToEmptyMap(t *testing.T) {

	prep := EmptySchema.Prepare()
	exp := map[string]SchemaField{}

	AssertStringsEq(t, prep.version, "4.0")
	AssertMapsEq(t, prep.hmap, exp)
	AssertSlicesEq(t, prep.requiredFields, []string{})
}

type Empty struct{}

var EmptySchema = NewSchema[Empty]("4.0")
