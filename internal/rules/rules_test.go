package rules

import "testing"

func TestDefaultRulesAreActionable(t *testing.T) {
	for _, hypothesis := range Default() {
		if hypothesis.Source == "" || hypothesis.Risk == "" {
			t.Fatalf("incomplete hypothesis %#v", hypothesis)
		}
		if hypothesis.Option != "" && hypothesis.Undo == "" {
			t.Fatalf("option without undo %#v", hypothesis)
		}
	}
}
