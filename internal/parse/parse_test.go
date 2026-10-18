package parse

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSpan(t *testing.T) {
	tests := map[string]time.Duration{
		"345ms":       345 * time.Millisecond,
		"1.234s":      1234 * time.Millisecond,
		"1min 2.345s": time.Minute + 2345*time.Millisecond,
		"2h 3min 4s":  2*time.Hour + 3*time.Minute + 4*time.Second,
		"1d 2h":       26 * time.Hour,
		"123us":       123 * time.Microsecond,
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := Span(input)
			if err != nil || got != want {
				t.Fatalf("Span(%q) = %v, %v; want %v, nil", input, got, err, want)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	input := "Startup finished in 2.001s (kernel) + 1.500s (userspace) = 3.501s.\n" +
		"graphical.target reached after 1min 2.345s in userspace.\n"
	got, err := ParseTime(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if got.Firmware != 0 || got.Loader != 0 || got.Initrd != 0 || got.Kernel != 2001*time.Millisecond || got.Userspace != 1500*time.Millisecond || got.Total != 3501*time.Millisecond || got.Target != "graphical.target" || got.TargetAfter != time.Minute+2345*time.Millisecond {
		t.Fatalf("ParseTime() = %#v", got)
	}
}

func TestParseTimeInProgress(t *testing.T) {
	_, err := ParseTime(strings.NewReader("Bootup is not yet finished. Please try again later.\n"))
	if !errors.Is(err, ErrBootInProgress) {
		t.Fatalf("ParseTime() error = %v; want ErrBootInProgress", err)
	}
}

func TestParseChain(t *testing.T) {
	input := "graphical.target @1min 2.345s\n" +
		"└─multi-user.target @1min 2.300s\n" +
		"  └─slow.service @2.000s +1min 0.300s\n"
	nodes, err := ParseChain(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 3 || nodes[0].Depth != 0 || nodes[1].Depth != 1 || nodes[2].Depth != 2 || nodes[2].Took != time.Minute+300*time.Millisecond {
		t.Fatalf("ParseChain() = %#v", nodes)
	}
}

func TestParseBlame(t *testing.T) {
	entries, err := ParseBlame(strings.NewReader("1min 2.345s slow.service\n345ms quick.service\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Unit != "slow.service" || entries[0].Took != time.Minute+2345*time.Millisecond || entries[1].Unit != "quick.service" {
		t.Fatalf("ParseBlame() = %#v", entries)
	}
}

func FuzzSpan(f *testing.F) {
	for _, value := range []string{"1s", "1min 2.3s", "bad", ""} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		_, _ = Span(value)
	})
}

func FuzzChainLine(f *testing.F) {
	for _, value := range []string{"foo.service @1s +2s", "└─foo.service", "bad"} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		_, _ = ParseChain(strings.NewReader(value))
	})
}

func TestParseJournal(t *testing.T) {
	input := strings.Join([]string{
		`{"__MONOTONIC_TIMESTAMP":"1000","UNIT":"first.service","MESSAGE":"first","PRIORITY":"6"}`,
		`not json`,
		`{"__MONOTONIC_TIMESTAMP":2000,"_SYSTEMD_UNIT":"second.service","MESSAGE":[255,254],"PRIORITY":3}`,
		`{"__MONOTONIC_TIMESTAMP":"3000","MESSAGE":"later"}`,
	}, "\n")
	entries, malformed, err := ParseJournal(strings.NewReader(input), 2*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if malformed != 1 || len(entries) != 2 || entries[0].Unit != "first.service" || entries[1].Unit != "second.service" || entries[1].Msg != string([]byte{255, 254}) || entries[1].Priority != 3 {
		t.Fatalf("ParseJournal() = %#v, %d", entries, malformed)
	}
}

func FuzzJournalLine(f *testing.F) {
	for _, value := range []string{"{}", `{"MESSAGE":"hello"}`, "bad"} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		_, _, _ = ParseJournal(strings.NewReader(value), 0)
	})
}
