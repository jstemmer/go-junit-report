package subtests

import "testing"

func TestSubtests(t *testing.T) {
	t.Run("Subtest", func(t *testing.T) {
		t.Log("ok")
	})
	t.Run("Subtest", func(t *testing.T) {
		t.Error("error message")
	})
	t.Run("Subtest", func(t *testing.T) {
		t.Skip("skip message")
	})
}

func TestNestedSubtests(t *testing.T) {
	t.Run("a#1", func(t *testing.T) {
		t.Run("b#1", func(t *testing.T) {
			t.Run("c#1", func(t *testing.T) {
			})
		})
	})
}

func TestFailingSubtestWithNestedSubtest(t *testing.T) {
	t.Run("Subtest", func(t *testing.T) {
		t.Run("Subsubtest", func(t *testing.T) {
			t.Log("ok")
		})
		t.Errorf("Subtest error message")
	})
}
