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

func FuzzSpan(f *testing.F) {
	for _, value := range []string{"1s", "1min 2.3s", "bad", ""} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		_, _ = Span(value)
	})
}
