package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lone4alker/bootwhy/internal/analyze"
	"github.com/lone4alker/bootwhy/internal/parse"
)

func TestPrintReport(t *testing.T) {
	var output bytes.Buffer
	printReport(&output, parse.Times{Kernel: time.Second, Userspace: 2 * time.Second, Total: 3 * time.Second, Target: "graphical.target", TargetAfter: 2 * time.Second}, []parse.ChainNode{{Unit: "graphical.target", At: 2 * time.Second}, {Unit: "slow.service", At: time.Second, Took: time.Second, Depth: 1}}, analyze.Result{})
	text := output.String()
	for _, want := range []string{"Boot phases:", "Total settled:", "graphical.target reached:", "Critical chain:", "slow.service @1s +1s"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q:\n%s", want, text)
		}
	}
}

func TestCLIIntegration(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "bootwhy")
	buildCmd := exec.Command("go", "build", "-o", binary, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, string(out))
	}

	// 1. --version
	out, err := exec.Command(binary, "--version").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "dev") {
		t.Fatalf("--version failed: %v, out: %s", err, string(out))
	}

	// 2. Clean boot (fedora) exits 0
	cleanCmd := exec.Command(binary, "--from", filepath.Join("..", "..", "testdata", "fedora"))
	cleanOut, err := cleanCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("clean fedora failed: %v, out: %s", err, string(cleanOut))
	}
	if !strings.Contains(string(cleanOut), "normal for this kind of machine") {
		t.Fatalf("clean fedora unexpected output: %s", string(cleanOut))
	}

	// 3. Fault boot (device-timeout) exits 1
	faultCmd := exec.Command(binary, "--from", filepath.Join("..", "..", "testdata", "faults", "device-timeout"))
	faultOut, err := faultCmd.CombinedOutput()
	if err == nil {
		t.Fatalf("device-timeout should exit 1 on delay detection, got exit 0")
	}
	if !strings.Contains(string(faultOut), "device-timeout") {
		t.Fatalf("expected device-timeout finding: %s", string(faultOut))
	}

	// 4. Explain 1 exits 0 and explains finding
	explainCmd := exec.Command(binary, "--from", filepath.Join("..", "..", "testdata", "faults", "device-timeout"), "explain", "1")
	explainOut, err := explainCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("explain 1 failed: %v, out: %s", err, string(explainOut))
	}
	if !strings.Contains(string(explainOut), "A mount or swap device did not appear in time") {
		t.Fatalf("unexpected explain output: %s", string(explainOut))
	}

	// 5. JSON output is valid JSON
	jsonCmd := exec.Command(binary, "--from", filepath.Join("..", "..", "testdata", "faults", "device-timeout"), "--json")
	jsonOut, _ := jsonCmd.CombinedOutput()
	if !json.Valid(jsonOut) {
		t.Fatalf("invalid json output: %s", string(jsonOut))
	}
}
