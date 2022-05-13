package race

import "testing"

func TestRace(t *testing.T) {
	done := make(chan bool)
	x := 0
	go func() {
		x = 5
		done <- true
	}()
	x = 3
	t.Logf("x = %d\n", x)
	<-done
}
