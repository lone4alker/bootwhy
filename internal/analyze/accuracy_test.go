package analyze

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lone4alker/bootwhy/internal/acquire"
	"github.com/lone4alker/bootwhy/internal/parse"
)

func TestFixtureShapes(t *testing.T) {
	directories, err := filepath.Glob(filepath.Join("..", "..", "testdata", "**", "expect.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, expectPath := range directories {
		expect, err := os.ReadFile(expectPath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(expect), "shape=") {
			t.Fatalf("%s has no shape", expectPath)
		}
		dir := filepath.Dir(expectPath)
		raw, err := acquire.Load(t.Context(), dir)
		if err != nil {
			t.Fatal(err)
		}
		times, _ := parse.ParseTime(strings.NewReader(string(raw.Time)))
		chain, _ := parse.ParseChain(strings.NewReader(string(raw.Chain)))
		blame, _ := parse.ParseBlame(strings.NewReader(string(raw.Blame)))
		got := Analyze(times, chain, blame, raw.Ctx)
		if strings.Contains(string(expect), "shape=normal") && got.Shape != Normal {
			t.Fatalf("%s: got %v", dir, got.Shape)
		}
	}
}
