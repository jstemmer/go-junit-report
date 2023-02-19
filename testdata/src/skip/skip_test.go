package skip

import "testing"

func TestSkip(t *testing.T) {
	t.Skip("skip message")
}

func TestSkipNow(t *testing.T) {
	t.Log("log message")
	t.SkipNow()
}

func TestSkipNoMessage(t *testing.T) {
	t.SkipNow()
}

func TestSkipLogMessage(t *testing.T) {
	t.Log("log message")
	t.Skip("skip message")
}
