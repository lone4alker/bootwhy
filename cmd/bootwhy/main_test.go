package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/lone4alker/bootwhy/internal/parse"
)

func TestPrintReport(t *testing.T) {
	var output bytes.Buffer
	printReport(&output, parse.Times{Kernel: time.Second, Userspace: 2 * time.Second, Total: 3 * time.Second, Target: "graphical.target", TargetAfter: 2 * time.Second}, []parse.ChainNode{{Unit: "graphical.target", At: 2 * time.Second}, {Unit: "slow.service", At: time.Second, Took: time.Second, Depth: 1}})
	text := output.String()
	for _, want := range []string{"Boot phases:", "Total settled:", "graphical.target reached:", "Critical chain:", "slow.service @1s +1s"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q:\n%s", want, text)
		}
	}
}
