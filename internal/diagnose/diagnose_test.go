package diagnose

import (
	"testing"

	"github.com/lone4alker/bootwhy/internal/acquire"
	"github.com/lone4alker/bootwhy/internal/analyze"
	"github.com/lone4alker/bootwhy/internal/parse"
	"github.com/lone4alker/bootwhy/internal/rules"
)

func TestDiagnoseConfidenceAndOption(t *testing.T) {
	hypothesis := rules.Hypothesis{Name: "failure", Option: "check status", Undo: "none", Risk: "safe", Source: "systemctl(1)", Evidence: []rules.Check{{Name: "failed", Weight: 10, Stated: true, Test: func(evidence *rules.Evidence) (bool, string) { return len(evidence.Entries) > 0, "journal: failed" }}}}
	contributors := []analyze.Contributor{{Unit: "bad.service"}}
	journal := []parse.Entry{{Unit: "bad.service", Msg: "failed"}}
	findings := Diagnose(contributors, nil, journal, acquire.Context{}, []rules.Hypothesis{hypothesis})
	if len(findings) != 1 || findings[0].Confidence != "high" {
		t.Fatalf("Diagnose() = %#v", findings)
	}
	low := Diagnose(contributors, nil, nil, acquire.Context{}, []rules.Hypothesis{hypothesis})
	if len(low) != 0 {
		t.Fatalf("Diagnose() without evidence = %#v", low)
	}
}
