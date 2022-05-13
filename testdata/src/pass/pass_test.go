package pass

import "testing"

func TestPass(t *testing.T) {
}

func TestPassLog(t *testing.T) {
	t.Logf("log %s", message())
	t.Log("log\nmulti\nline")
}
