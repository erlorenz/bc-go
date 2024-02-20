package bc

import (
	"testing"
)

type TestNameGotWant[T comparable] struct {
	name string
	got  T
	want T
}

func RunTable[T comparable](table []TestNameGotWant[T], t *testing.T) {
	for _, v := range table {
		t.Run(v.name, func(t *testing.T) {
			if v.got != v.want {
				t.Errorf("error: got %v, wanted %v", v.got, v.want)
			}
		})
	}
}
