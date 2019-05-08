package parser

import (
	"strings"
	"testing"
	"time"
)

func TestParseSeconds(t *testing.T) {
	tests := []struct {
		in string
		d  time.Duration
	}{
		{"", 0},
		{"4", 4 * time.Second},
		{"0.1", 100 * time.Millisecond},
		{"0.050", 50 * time.Millisecond},
		{"2.003", 2*time.Second + 3*time.Millisecond},
	}

	for _, test := range tests {
		d := parseSeconds(test.in)
		if d != test.d {
			t.Errorf("parseSeconds(%q) == %v, want %v\n", test.in, d, test.d)
		}
	}
}

func TestParseNanoseconds(t *testing.T) {
	tests := []struct {
		in string
		d  time.Duration
	}{
		{"", 0},
		{"0.1", 0 * time.Nanosecond},
		{"0.9", 0 * time.Nanosecond},
		{"4", 4 * time.Nanosecond},
		{"5000", 5 * time.Microsecond},
		{"2000003", 2*time.Millisecond + 3*time.Nanosecond},
	}

	for _, test := range tests {
		d := parseNanoseconds(test.in)
		if d != test.d {
			t.Errorf("parseSeconds(%q) == %v, want %v\n", test.in, d, test.d)
		}
	}
}

func TestParseFailePackage(t *testing.T) {
	reader := strings.NewReader(
	"=== RUN   TestGetMixerSAN\n" +
	"--- PASS: TestGetMixerSAN (0.00s)\n" +
	"=== RUN   TestGetPilotSAN\n" +
	"--- PASS: TestGetPilotSAN (0.00s)\n" +
	"--- PASS: TestGetPilotSAN (0.00s)\n" +
	"=== RUN TestApplyThrice\n" +
	"2019-05-02T02:00:14.273826Z	info	Graceful termination period is -10s, starting...\n" +
	"--- PASS: TestApplyThrice (0.00s)\n" +
	"2019-05-02T02:00:14.274025Z	info	Epoch 0 starting\n" +
	"panic: Fail in goroutine after TestApplyThrice has completed\n" +
	"FAIL	fake_test	0.008s\n")
	res, _ := Parse(reader, "fake_test")
	if len(res.Packages) != 1 {
		t.Errorf("We have only one package.")
	}
        if len(res.Packages[0].Tests) != 3 {
                t.Errorf("We expect to parse 3 tests.")
        }
	if res.Packages[0].Tests[0].Result != 0 || res.Packages[0].Tests[1].Result != 0 {
		t.Errorf("We expect first two tests to Pass.")
	}
	if res.Packages[0].Tests[2].Result != 1 {
		t.Errorf("We expect third test to Fail.")
	}
}

func TestParseAllPass(t *testing.T) {
        reader := strings.NewReader(
        "=== RUN   TestGetMixerSAN\n" +
        "--- PASS: TestGetMixerSAN (0.00s)\n" +
        "=== RUN   TestGetPilotSAN\n" +
        "--- PASS: TestGetPilotSAN (0.00s)\n" +
        "--- PASS: TestGetPilotSAN (0.00s)\n" +
        "=== RUN TestApplyThrice\n" +
        "2019-05-02T02:00:14.273826Z    info    Graceful termination period is -10s, starting...\n" +
        "--- PASS: TestApplyThrice (0.00s)\n" +
        "PASS\n")
        res, _ := Parse(reader, "fake_test")
        if len(res.Packages) != 1 {
                t.Errorf("We have only one package.")
        }
        if len(res.Packages[0].Tests) != 3 {
                t.Errorf("We expect to parse 3 tests.")
        }
        if res.Packages[0].Tests[0].Result != 0 || res.Packages[0].Tests[1].Result != 0 || res.Packages[0].Tests[2].Result != 0 {
                t.Errorf("We expect all tests to Pass.")
        }
}

