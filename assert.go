package vcard

import (
	"errors"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func assertEq[T comparable](t *testing.T, found T, expected T) {
	if found != expected {
		t.Errorf("Values are different.\nExpected:\n\n%v\n\nFound:\n\n%v", expected, found)
	}
}

func assertSlicesEq[T comparable](t *testing.T, found []T, expected []T) {
	if !slices.Equal(found, expected) {
		t.Errorf("Slices are different.\nExpected:\n\n%v\n\nFound:\n\n%v", expected, found)
	}
}

func assertStringsEq(t *testing.T, found string, expected string) {
	if found != expected {
		t.Errorf("Strings are different.\nExpected:\n\n%s\n\nFound:\n\n%s", expected, found)
	}
}

func assertStringLinesEq(t *testing.T, found string, expected string) {

	foundCountMap := map[string]int{}
	for line := range strings.Lines(found) {
		foundCountMap[line]++
	}
	expectedCountMap := map[string]int{}
	for line := range strings.Lines(expected) {
		expectedCountMap[line]++
	}

	if !maps.Equal(foundCountMap, expectedCountMap) {
		t.Errorf("Strings are different.\nExpected len=%v:\n\n%s\nFound len=%v:\n\n%s", len(expected), expected, len(found), found)
	}
}

func assertMapsEq[K comparable, V comparable](t *testing.T, found map[K]V, expected map[K]V) {
	if !maps.Equal(found, expected) {
		t.Errorf("Maps are different.\nExpected:\n\n%v\n\nFound:\n\n%v", expected, found)
	}
}

func assertErrIs(t *testing.T, found error, expected error, substr string) {
	if !errors.Is(found, expected) {
		t.Errorf("Error has wrong type %q but %q was expected", reflect.TypeOf(found), reflect.TypeOf(expected))
	}
	if !strings.Contains(found.Error(), substr) {
		t.Errorf("Error has right type %q but description does not contain %q", reflect.TypeOf(found), substr)
	}
}

func assertStringContains(t *testing.T, s string, substr string) {
	if !strings.Contains(s, substr) {
		t.Errorf("Failed to find substring %q in:\n%s", substr, s)
	}
}

// Transforms LF-strings ("\n" as newline character) into CRLF-strings ("\r\n" as newline sequence)
// This is only used in testing to simplify definition of expected value.
func crlfy(s string) string {

	buf := strings.Builder{}

	for line := range strings.Lines(s) {
		s := strings.Trim(line, " \r\n")
		if s == "" {
			continue
		}
		buf.WriteString(s)
		buf.WriteString("\r\n")
	}

	return buf.String()
}
