package pkg

import (
	"testing"
	"time"
)

func TestP1(t *testing.T) {
	t.Parallel()
	t.Log("t.Log(P1)")
	time.Sleep(100 * time.Millisecond)
	t.Errorf("P1 error")
}

func TestP2(t *testing.T) {
	t.Parallel()
	t.Log("t.Log(P2)")
	time.Sleep(50 * time.Millisecond)
	t.Errorf("P2 error")
}

func TestP3(t *testing.T) {
	t.Parallel()
	t.Log("t.Log(P3)")
	time.Sleep(75 * time.Millisecond)
	t.Errorf("P3 error")
}
