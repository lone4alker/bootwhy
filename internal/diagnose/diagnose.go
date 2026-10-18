package diagnose

import (
	"sort"
	"strings"

	"github.com/lone4alker/bootwhy/internal/acquire"
	"github.com/lone4alker/bootwhy/internal/analyze"
	"github.com/lone4alker/bootwhy/internal/parse"
	"github.com/lone4alker/bootwhy/internal/rules"
)

type Scored struct {
	Hypothesis rules.Hypothesis
	Score      int
	Evidence   []string
}

type Finding struct {
	analyze.Contributor
	Ranked     []Scored
	Confidence string
	Tied       bool
	Clock      string
}

const tieMargin = 0.20

// Diagnose ranks hypotheses for each contributor using its journal evidence.
func Diagnose(contributors []analyze.Contributor, chain []parse.ChainNode, journal []parse.Entry, context acquire.Context, hypotheses []rules.Hypothesis) []Finding {
	findings := make([]Finding, 0, len(contributors))
	for _, contributor := range contributors {
		evidence := collectEvidence(contributor.Unit, chain, journal, context)
		ranked, stated := rank(&evidence, hypotheses)
		if len(ranked) == 0 {
			continue
		}
		finding := Finding{Contributor: contributor, Ranked: ranked, Clock: "desktop"}
		finding.Tied = len(ranked) > 1 && ranked[0].Score-ranked[1].Score <= int(float64(ranked[0].Score)*tieMargin)
		finding.Confidence = confidence(stated, evidence, finding.Tied)
		if finding.Confidence == "low" {
			finding.Ranked[0].Hypothesis.Option = ""
		}
		findings = append(findings, finding)
	}
	return findings
}

func collectEvidence(unit string, chain []parse.ChainNode, journal []parse.Entry, context acquire.Context) rules.Evidence {
	evidence := rules.Evidence{Unit: unit, JournalMissing: len(journal) == 0, JournalPartial: context.JournalPartial}
	for index, node := range chain {
		if node.Unit == unit && index > 0 {
			evidence.PreviousUnit = chain[index-1].Unit
		}
		if node.Unit != unit && node.Took > 0 {
			evidence.Overlapping = append(evidence.Overlapping, node.Unit)
		}
	}
	for _, entry := range journal {
		if entry.Unit == unit {
			evidence.Entries = append(evidence.Entries, entry)
		}
	}
	return evidence
}

func rank(evidence *rules.Evidence, hypotheses []rules.Hypothesis) ([]Scored, bool) {
	var ranked []Scored
	stated := false
	for _, hypothesis := range hypotheses {
		if !applies(hypothesis, evidence.Unit) {
			continue
		}
		scored := Scored{Hypothesis: hypothesis}
		for _, check := range hypothesis.Evidence {
			if hit, line := check.Test(evidence); hit {
				scored.Score += check.Weight
				scored.Evidence = append(scored.Evidence, line)
				stated = stated || check.Stated
			}
		}
		if scored.Score > 0 {
			ranked = append(ranked, scored)
		}
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].Score > ranked[j].Score })
	return ranked, stated
}

func applies(hypothesis rules.Hypothesis, unit string) bool {
	for _, pattern := range hypothesis.Units {
		if strings.Contains(unit, pattern) {
			return true
		}
	}
	return len(hypothesis.Units) == 0
}

func confidence(stated bool, evidence rules.Evidence, tied bool) string {
	if evidence.JournalMissing || tied {
		return "low"
	}
	if stated && !evidence.JournalPartial {
		return "high"
	}
	return "medium"
}
