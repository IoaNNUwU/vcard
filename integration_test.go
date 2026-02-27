package vcard

import (
	"io"
	"os"
	"testing"
)

func TestMultipleEncodesProduceSameResultMap(t *testing.T) {
	f, err := os.Open("assets/vCardv4.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	m1 := make(map[string]string)
	Unmarshal(data, &m1)
	b1, err := Marshal(m1)

	m2 := make(map[string]string)
	Unmarshal(b1, &m2)
	b2, err := Marshal(m2)

	assertMapsEq(t, m1, m2)
	assertStringLinesEq(t, string(b1), string(b2))
}

func TestMultipleEncodesProduceSameResultSliceOfOneMap(t *testing.T) {
	f, err := os.Open("assets/vCardv4.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	s1 := make([]map[string]string, 0)
	Unmarshal(data, &s1)
	b1, err := Marshal(s1)

	s2 := make([]map[string]string, 0)
	Unmarshal(b1, &s2)
	b2, err := Marshal(s2)

	assertMapsEq(t, s1[0], s2[0])
	assertStringLinesEq(t, string(b1), string(b2))
}

func TestMultipleEncodesProduceSameResultSliceOfManyMap(t *testing.T) {
	f, err := os.Open("assets/vCardsv4.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	s1 := make([]map[string]string, 0)
	Unmarshal(data, &s1)
	b1, err := Marshal(s1)

	s2 := make([]map[string]string, 0)
	Unmarshal(b1, &s2)
	b2, err := Marshal(s2)

	for i := range 8 {
		assertMapsEq(t, s1[i], s2[i])
	}
	assertStringLinesEq(t, string(b1), string(b2))
}

type customSerializableField struct {
	s string
}

func (csf *customSerializableField) MarshalVCardField() ([]byte, error) {
	return []byte(csf.s), nil
}

func (csf *customSerializableField) UnmarshalVCardField(data []byte) error {
	csf.s = string(data)
	return nil
}

func TestMultipleEncodesProduceSameResultMapMarshalerValue(t *testing.T) {
	f, err := os.Open("assets/vCardv4.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	m1 := make(map[string]customSerializableField)
	Unmarshal(data, &m1)
	b1, err := Marshal(m1)

	m2 := make(map[string]customSerializableField)
	Unmarshal(b1, &m2)
	b2, err := Marshal(m2)

	assertMapsEq(t, m1, m2)
	assertStringLinesEq(t, string(b1), string(b2))
}

func TestMultipleEncodesProduceSameResultSliceOfOneMapMarshalerValue(t *testing.T) {
	f, err := os.Open("assets/vCardv4.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	s1 := make([]map[string]customSerializableField, 0)
	Unmarshal(data, &s1)
	b1, err := Marshal(s1)

	s2 := make([]map[string]customSerializableField, 0)
	Unmarshal(b1, &s2)
	b2, err := Marshal(s2)

	assertMapsEq(t, s1[0], s2[0])
	assertStringLinesEq(t, string(b1), string(b2))
}

func TestMultipleEncodesProduceSameResultSliceOfManyMapMarshalerValue(t *testing.T) {
	f, err := os.Open("assets/vCardsv4.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	s1 := make([]map[string]customSerializableField, 0)
	Unmarshal(data, &s1)
	b1, err := Marshal(s1)

	s2 := make([]map[string]customSerializableField, 0)
	Unmarshal(b1, &s2)
	b2, err := Marshal(s2)

	for i := range 8 {
		assertMapsEq(t, s1[i], s2[i])
	}
	assertStringLinesEq(t, string(b1), string(b2))
}

type customUser struct {
	Name 		string `vCard:"N"`
	DisplayName string `vCard:"FN"`
	Desc 		string `vCard:"NAME"`
}

func TestMultipleEncodesProduceSameResultStruct(t *testing.T) {
	f, err := os.Open("assets/vCardv4.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	u1 := customUser{}
	Unmarshal(data, &u1)
	b1, err := Marshal(u1)

	u2 := customUser{}
	Unmarshal(b1, &u2)
	b2, err := Marshal(u2)

	assertEq(t, u1, u2)
	assertStringLinesEq(t, string(b1), string(b2))
}

func TestMultipleEncodesProduceSameResultSliceOfOneStruct(t *testing.T) {
	f, err := os.Open("assets/vCardv4.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	s1 := make([]customUser, 0)
	Unmarshal(data, &s1)
	b1, err := Marshal(s1)

	s2 := make([]customUser, 0)
	Unmarshal(b1, &s2)
	b2, err := Marshal(s2)

	assertEq(t, s1[0], s2[0])
	assertStringLinesEq(t, string(b1), string(b2))
}

func TestMultipleEncodesProduceSameResultSliceOfManyStructs(t *testing.T) {
	f, err := os.Open("assets/vCardsv4.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	s1 := make([]customUser, 0)
	Unmarshal(data, &s1)
	b1, err := Marshal(s1)

	s2 := make([]customUser, 0)
	Unmarshal(b1, &s2)
	b2, err := Marshal(s2)

	for i := range 8 {
		assertEq(t, s1[i], s2[i])
	}
	assertStringLinesEq(t, string(b1), string(b2))
}

type customUserMarshaler struct {
	Name 		customSerializableField `vCard:"N"`
	DisplayName customSerializableField `vCard:"FN"`
	Desc 		customSerializableField `vCard:"NAME"`
}

func TestMultipleEncodesProduceSameResultStructMarshalerField(t *testing.T) {
	f, err := os.Open("assets/vCardv4.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	m1 := customUserMarshaler{}
	Unmarshal(data, &m1)
	b1, err := Marshal(m1)

	m2 := customUserMarshaler{}
	Unmarshal(b1, &m2)
	b2, err := Marshal(m2)

	assertEq(t, m1, m2)
	assertStringLinesEq(t, string(b1), string(b2))
}

func TestMultipleEncodesProduceSameResultSliceOfOneStructMarshalerField(t *testing.T) {
	f, err := os.Open("assets/vCardv4.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	s1 := make([]customUserMarshaler, 0)
	Unmarshal(data, &s1)
	b1, err := Marshal(s1)

	s2 := make([]customUserMarshaler, 0)
	Unmarshal(b1, &s2)
	b2, err := Marshal(s2)

	assertEq(t, s1[0], s2[0])
	assertStringLinesEq(t, string(b1), string(b2))
}

func TestMultipleEncodesProduceSameResultSliceOfManyStructsMarshalerField(t *testing.T) {
	f, err := os.Open("assets/vCardsv4.vcf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	s1 := make([]customUserMarshaler, 0)
	Unmarshal(data, &s1)
	b1, err := Marshal(s1)

	s2 := make([]customUserMarshaler, 0)
	Unmarshal(b1, &s2)
	b2, err := Marshal(s2)

	for i := range 8 {
		assertEq(t, s1[i], s2[i])
	}
	assertStringLinesEq(t, string(b1), string(b2))
}