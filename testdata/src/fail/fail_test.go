package fail

import "testing"

func TestOne(t *testing.T) {
	t.Errorf("Error message")
	t.Errorf("Longer\nerror\nmessage.")
}
