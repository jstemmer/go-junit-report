package skip

import "testing"

func TestSkip(t *testing.T) {
	t.Skip("skip message")
}

func TestSkipNow(t *testing.T) {
	t.Log("log message")
	t.SkipNow()
}
