package vcard

import (
	"io"
	"os"
	"testing"
)

func TestMultipleEncodesProduceSameResult(t *testing.T) {
	f, err := os.Open("assets/vCard.vcf")
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

	_ = b2

	assertMapsEq(t, s1[0], s2[0])

	// TODO: s2 LOST URL field
}