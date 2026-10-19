package analyze_test

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lone4alker/bootwhy/internal/acquire"
	"github.com/lone4alker/bootwhy/internal/analyze"
	"github.com/lone4alker/bootwhy/internal/diagnose"
	"github.com/lone4alker/bootwhy/internal/parse"
	"github.com/lone4alker/bootwhy/internal/rules"
)

func TestFixtureShapes(t *testing.T) {
	testdataRoot := filepath.Join("..", "..", "testdata")
	var expectPaths []string
	err := filepath.WalkDir(testdataRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "expect.txt" {
			expectPaths = append(expectPaths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(expectPaths) == 0 {
		t.Fatal("no expect.txt fixtures found")
	}

	for _, expectPath := range expectPaths {
		dir := filepath.Dir(expectPath)
		t.Run(filepath.ToSlash(dir), func(t *testing.T) {
			expectBytes, err := os.ReadFile(expectPath)
			if err != nil {
				t.Fatal(err)
			}
			expect := string(expectBytes)

			raw, err := acquire.Load(t.Context(), dir)
			if err != nil {
				t.Fatal(err)
			}

			times, err := parse.ParseTime(bytes.NewReader(raw.Time))
			if strings.Contains(expect, "in_progress=true") {
				if !errors.Is(err, parse.ErrBootInProgress) {
					t.Fatalf("%s: expected ErrBootInProgress, got %v", dir, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("%s: unexpected error parsing time: %v", dir, err)
			}

			chain, _ := parse.ParseChain(bytes.NewReader(raw.Chain))
			blame, _ := parse.ParseBlame(bytes.NewReader(raw.Blame))
			got := analyze.Analyze(times, chain, blame, raw.Ctx)

			if strings.Contains(expect, "shape=normal") && got.Shape != analyze.Normal {
				t.Fatalf("%s: got shape %v; want Normal", dir, got.Shape)
			}
			if strings.Contains(expect, "shape=single") && got.Shape != analyze.Single {
				t.Fatalf("%s: got shape %v; want Single", dir, got.Shape)
			}

			for _, line := range strings.Split(expect, "\n") {
				key, val, ok := strings.Cut(strings.TrimSpace(line), "=")
				if !ok || key != "cause" {
					continue
				}
				expectedCause := val
				journal, _, _ := parse.ParseJournal(bytes.NewReader(raw.Journal), times.Kernel+times.Initrd+times.TargetAfter+1000000000)
				findings := diagnose.Diagnose(got.Contributors, chain, journal, raw.Ctx, rules.Default())
				if len(findings) == 0 {
					t.Fatalf("%s: expected findings for cause %q, got none", dir, expectedCause)
				}
				topHypothesis := ""
				if len(findings[0].Ranked) > 0 {
					topHypothesis = findings[0].Ranked[0].Hypothesis.Name
				}
				if topHypothesis != expectedCause {
					t.Fatalf("%s: top diagnosed cause = %q; want %q", dir, topHypothesis, expectedCause)
				}
			}
		})
	}
}
