package main

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2/test"
)

func TestGreeting(t *testing.T) {
	out, in := makeUI()

	if out.Text != "Hello world!" {
		t.Error("Incorrect initial greeting")
	}

	test.Type(in, "Andy")
	if out.Text != "Hello Andy!" {
		t.Error("Incorrect user greeting")
	}
}
