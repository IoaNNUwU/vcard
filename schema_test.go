package vcard

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

	assertStringsEq(t, schema.version, exp.version)
	assertMapsEq(t, schema.fields, exp.fields)
	assertMapsEq(t, schema.requiredFields, exp.requiredFields)
}
