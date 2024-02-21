package slicefuncs

import (
	"strings"
	"testing"
)

func TestFilter(t *testing.T) {
	names := []string{"john", "lisa", "fred", "james"}

	compare := func(name string) bool {
		return strings.HasPrefix(name, "j")
	}

	newNames := Filter(names, compare)

	if len(newNames) != 2 {
		t.Errorf("wanted 2, got %d: %v ", len(newNames), newNames)
	}
}

func TestMap(t *testing.T) {
	names := []string{"john"}

	transform := func(name string) string {
		return strings.ToUpper(name)
	}

	newNames := Map(names, transform)

	if newNames[0] != "JOHN" {
		t.Errorf("wanted JOHN, got %s", newNames[0])
	}
}
