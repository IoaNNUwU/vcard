package vcardgo

import "testing"

type TestImplementation struct {
	N    string
	NAME string `vCard:"required"`
	FN   string `vCard:"required"`
}

func TestSchemaToMap(t *testing.T) {

	schema := SchemaFor[TestImplementation]("3.0")

	exp := Schema{
		version: "3.0",
		fields: map[string]struct{}{
			"N":    {},
			"NAME": {},
			"FN":   {},
		},
		requiredFields: map[string]struct{}{
			"NAME": {},
			"FN":   {},
		},
	}

	AssertStringsEq(t, schema.version, exp.version)
	AssertMapsEq(t, schema.fields, exp.fields)
	AssertMapsEq(t, schema.requiredFields, exp.requiredFields)
}
