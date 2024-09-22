package fns_test

import (
	"testing"

	"github.com/yaoshangnetwork/gobase/fns"
)

func TestNilOr(t *testing.T) {
	var a *string
	if fns.NilOr(a, "def") != "def" {
		t.Error("err")
	}

	c := "123"
	var b *string = &c
	if fns.NilOr(b, "def") != c {
		t.Error("err")
	}
}
