package main

import (
	"fmt"
	"testing"
)

func TestFlat(t *testing.T) {
	t.Logf("log 1")
	t.Logf("log 2")
	fmt.Printf("printf 1\n")
	fmt.Printf("printf 2\n")
}

func TestWithSpace(t *testing.T) {
	t.Logf("no-space")
	t.Logf(" one-space")
	t.Logf("  two-space")
	t.Logf("    four-space")
	t.Logf("        eight-space")
	t.Logf("no-space")
	fmt.Printf("no-space\n")
	fmt.Printf(" one-space\n")
	fmt.Printf("  two-space\n")
	fmt.Printf("    four-space\n")
	fmt.Printf("        eight-space\n")
	fmt.Printf("no-space\n")
}

func TestWithTab(t *testing.T) {
	t.Logf("no-tab")
	t.Logf("\tone-tab")
	t.Logf("\t\ttwo-tab")
	fmt.Printf("no-tab\n")
	fmt.Printf("\tone-tab\n")
	fmt.Printf("\t\ttwo-tab\n")
}

func TestWithNewlinesFlat(t *testing.T) {
	t.Logf("no-newline")
	t.Logf("one-newline\none-newline")
	t.Logf("two-newlines\ntwo-newlines\ntwo-newlines")
	fmt.Println("no-newline")
	fmt.Println("one-newline\none-newline")
	fmt.Println("two-newlines\ntwo-newlines\ntwo-newlines")
}

func TestSubTests(t *testing.T) {
	t.Run("TestFlat", TestFlat)
	t.Run("TestWithSpace", TestWithSpace)
	t.Run("TestWithTab", TestWithTab)
	t.Run("TestWithNewlinesFlat", TestWithNewlinesFlat)
}
