package vcard

import (
	"maps"
	"slices"
	"strings"
	"testing"
)

func AssertEq[T comparable](t *testing.T, found T, expected T) {
	if found != expected {
		t.Errorf("Values are different.\nExpected:\n\n%v\n\nFound:\n\n%v", expected, found)
	}
}

func AssertSlicesEq[T comparable](t *testing.T, found []T, expected []T) {
	if !slices.Equal(found, expected) {
		t.Errorf("Slices are different.\nExpected:\n\n%v\n\nFound:\n\n%v", expected, found)
	}
}

func AssertStringsEq(t *testing.T, found string, expected string) {
	if found != expected {
		t.Errorf("Strings are different.\nExpected:\n\n%s\n\nFound:\n\n%s", expected, found)
	}
}

func AssertStringLinesEq(t *testing.T, found string, expected string) {

	foundCountMap := map[string]int{}
	for line := range strings.Lines(found) {
		foundCountMap[line]++
	}
	expectedCountMap := map[string]int{}
	for line := range strings.Lines(expected) {
		expectedCountMap[line]++
	}

	if !maps.Equal(foundCountMap, expectedCountMap) {
		t.Errorf("Strings are different.\nExpected:\n\n%s\n\nFound:\n\n%s", expected, found)
	}
}

func AssertMapsEq[K comparable, V comparable](t *testing.T, found map[K]V, expected map[K]V) {
	if !maps.Equal(found, expected) {
		t.Errorf("Maps are different.\nExpected:\n\n%v\n\nFound:\n\n%v", expected, found)
	}
}

func AssertErr(t *testing.T, err error) {
	if err == nil {
		t.Error("Expected error")
	}
}

func AssertStringContains(t *testing.T, s string, substr string) {
	if !strings.Contains(s, substr) {
		t.Errorf("Failed to find substring %q in:\n%s", substr, s)
	}
}
