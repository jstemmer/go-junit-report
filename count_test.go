package main

import (
	"math/rand"
	"testing"
)

func TestFoo(t *testing.T) {
	t.Log("foo")
	t.Log("bar")
	if rand.Int()%2 == 0 {
		t.Errorf("test failed due to even number rand")
	}
}
