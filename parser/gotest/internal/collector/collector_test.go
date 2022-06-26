package collector

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClear(t *testing.T) {
	o := New()
	o.Append(1, "1")
	o.Append(2, "2")
	o.Clear(1)

	want := []string(nil)
	got := o.Get(1)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Clear(1) did not clear output (-want +got):\n%s", diff)
	}

	want = []string{"2"}
	got = o.Get(2)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Clear(1) cleared wrong output (-want +got):\n%s", diff)
	}
}

func TestAppendAndGet(t *testing.T) {
	o := New()
	o.Append(1, "1.1")
	o.Append(1, "1.2")
	o.Append(2, "2")
	o.Append(1, "1.3")

	want := []string{"1.1", "1.2", "1.3"}
	got := o.Get(1)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Append() incorrect (-want +got):\n%s", diff)
	}
}

func TestContains(t *testing.T) {
	o := New()
	o.Append(1, "1")
	o.Append(2, "2")
	o.Clear(1)

	if !o.Contains(2) {
		t.Errorf("Contains(1) incorrect, got true want false")
	}
	for i := -100; i < 100; i++ {
		if i != 2 && o.Contains(i) {
			t.Errorf("Contains(%d) incorrect, got true want false", i)
		}
	}
}

func TestGetAll(t *testing.T) {
	o := New()
	for i := 1; i <= 10; i++ {
		o.Append(i%3, strconv.Itoa(i))
	}

	want := []string{"1", "2", "4", "5", "7", "8", "10"}
	got := o.GetAll(1, 2)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("GetAll(1, 2) incorrect (-want +got):\n%s", diff)
	}
}

func TestMerge(t *testing.T) {
	o := New()
	for i := 1; i <= 10; i++ {
		o.Append(i%3, strconv.Itoa(i))
	}

	o.Merge(2, 1)

	want := []string{"1", "2", "4", "5", "7", "8", "10"}
	got := o.Get(1)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Get(1) after Merge(2, 1) incorrect (-want +got):\n%s", diff)
	}

	want = []string(nil)
	got = o.Get(2)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Get(2) after Merge(2, 1) incorrect (-want +got):\n%s", diff)
	}
}
