package stdout

import (
	"fmt"
	"testing"
)

func TestFailWithStdoutAndTestOutput(t *testing.T) {
	fmt.Print("multi\nline\nstdout\n")
	fmt.Print("single-line stdout\n")
	t.Errorf("single-line error")
	t.Errorf("multi\nline\nerror")
}

func TestFailWithStdoutAndNoTestOutput(t *testing.T) {
	fmt.Print("multi\nline\nstdout\n")
	fmt.Print("single-line stdout\n")
	t.Fail()
}

func TestFailWithTestOutput(t *testing.T) {
	t.Errorf("single-line error")
	t.Errorf("multi\nline\nerror")
}

func TestFailWithNoTestOutput(t *testing.T) {
	t.Fail()
}

func TestPassWithStdoutAndTestOutput(t *testing.T) {
	fmt.Print("multi\nline\nstdout\n")
	fmt.Print("single-line stdout\n")
	t.Log("single-line info")
	t.Log("multi\nline\ninfo")
}

func TestPassWithStdoutAndNoTestOutput(t *testing.T) {
	fmt.Print("multi\nline\nstdout\n")
	fmt.Print("single-line stdout\n")
}

func TestPassWithTestOutput(t *testing.T) {
	t.Log("single-line info")
	t.Log("multi\nline\ninfo")
}

func TestPassWithNoTestOutput(t *testing.T) {
}

func TestSubtests(t *testing.T) {
	t.Run("TestFailWithStdoutAndTestOutput", TestFailWithStdoutAndTestOutput)
	t.Run("TestFailWithStdoutAndNoTestOutput", TestFailWithStdoutAndNoTestOutput)
	t.Run("TestFailWithTestOutput", TestFailWithTestOutput)
	t.Run("TestFailWithNoTestOutput", TestFailWithNoTestOutput)
	t.Run("TestPassWithStdoutAndTestOutput", TestPassWithStdoutAndTestOutput)
	t.Run("TestPassWithStdoutAndNoTestOutput", TestPassWithStdoutAndNoTestOutput)
	t.Run("TestPassWithTestOutput", TestPassWithTestOutput)
	t.Run("TestPassWithNoTestOutput", TestPassWithNoTestOutput)
}
