package analyze

import (
	"testing"
	"time"

	"github.com/lone4alker/bootwhy/internal/acquire"
	"github.com/lone4alker/bootwhy/internal/parse"
)

func TestAnalyzeSlowUnit(t *testing.T) {
	chain := []parse.ChainNode{{Unit: "a.service", Took: time.Second}, {Unit: "slow.service", Took: 10 * time.Second}, {Unit: "b.service", Took: time.Second}}
	result := Analyze(parse.Times{}, chain, []parse.BlameEntry{{Unit: "parallel.service", Took: 10 * time.Second}}, acquire.Context{})
	if result.Shape != Single || len(result.Contributors) != 1 || result.Contributors[0].Unit != "slow.service" || len(result.Noted) != 1 {
		t.Fatalf("Analyze() = %#v", result)
	}
}

func TestAnalyzeVMScaling(t *testing.T) {
	chain := []parse.ChainNode{{Unit: "a.service", Took: time.Second}, {Unit: "slow.service", Took: 10 * time.Second}, {Unit: "b.service", Took: time.Second}}
	bare := Analyze(parse.Times{}, chain, nil, acquire.Context{})
	vm := Analyze(parse.Times{}, chain, nil, acquire.Context{VM: true})
	if vm.Contributors[0].Added >= bare.Contributors[0].Added {
		t.Fatalf("VM added %s; bare added %s", vm.Contributors[0].Added, bare.Contributors[0].Added)
	}
}

func TestAnalyzeEmptyChain(t *testing.T) {
	if got := Analyze(parse.Times{}, nil, nil, acquire.Context{}); got.Shape != MissingData {
		t.Fatalf("Analyze() = %#v", got)
	}
}
