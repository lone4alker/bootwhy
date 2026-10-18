package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/lone4alker/bootwhy/internal/analyze"
	"github.com/lone4alker/bootwhy/internal/diagnose"
	"github.com/lone4alker/bootwhy/internal/parse"
)

func Table(writer io.Writer, times parse.Times, result analyze.Result, findings []diagnose.Finding, source string) {
	label := "Boot to desktop"
	if strings.Contains(times.Target, "multi-user") {
		label = "Boot to login prompt"
	}
	if result.Shape == analyze.Normal {
		fmt.Fprintf(writer, "%s: %s    normal for this kind of machine\n", label, bootClock(times))
		return
	}
	fmt.Fprintf(writer, "%s: %s    total settled: %s\n", label, bootClock(times), times.Total)
	fmt.Fprintf(writer, "(from: %s)\n", source)
	fmt.Fprintln(writer, "#  Unit                     Added    Share  Likely cause              Confidence")
	for index, contributor := range result.Contributors {
		cause, confidence := "slow work", "medium"
		if index < len(findings) && len(findings[index].Ranked) > 0 {
			cause, confidence = findings[index].Ranked[0].Hypothesis.Name, findings[index].Confidence
		}
		fmt.Fprintf(writer, "%d  %-24s %-8s %3.0f%%  %-25s %s\n", index+1, middle(contributor.Unit, 24), contributor.Added, contributor.Share*100, cause, confidence)
	}
	if result.Rest > 0 {
		fmt.Fprintf(writer, "   (everything else)         %s\n", result.Rest)
	}
	if result.Shape == analyze.Several {
		fmt.Fprintln(writer, "Fix row 1 first.")
	}
	fmt.Fprintln(writer, "Next: bootwhy explain 1")
}

func Explain(writer io.Writer, finding diagnose.Finding) {
	if len(finding.Ranked) == 0 {
		fmt.Fprintln(writer, "No diagnosis is available for this contributor.")
		return
	}
	best := finding.Ranked[0]
	fmt.Fprintf(writer, "%s: %s (%s confidence)\n", finding.Unit, best.Hypothesis.Explain, finding.Confidence)
	for _, line := range best.Evidence {
		fmt.Fprintf(writer, "Evidence: %s\n", line)
	}
	fmt.Fprintf(writer, "Source: %s\n", best.Hypothesis.Source)
	if best.Hypothesis.Option != "" {
		fmt.Fprintf(writer, "Option: %s\nRisk: %s\nUndo: %s\n", best.Hypothesis.Option, best.Hypothesis.Risk, best.Hypothesis.Undo)
	}
	if len(finding.Ranked) > 1 {
		fmt.Fprintf(writer, "Runner-up: %s\n", finding.Ranked[1].Hypothesis.Name)
	}
}

func JSON(writer io.Writer, times parse.Times, result analyze.Result, findings []diagnose.Finding) error {
	return json.NewEncoder(writer).Encode(struct {
		SchemaVersion int                `json:"schema_version"`
		Times         parse.Times        `json:"times"`
		Result        analyze.Result     `json:"result"`
		Findings      []diagnose.Finding `json:"findings"`
	}{1, times, result, findings})
}

func Timeline(writer io.Writer, times parse.Times, chain []parse.ChainNode) {
	longest := times.Total
	for _, node := range chain {
		if node.Took > longest {
			longest = node.Took
		}
	}
	if longest == 0 {
		return
	}
	fmt.Fprintln(writer, "Timeline:")
	for _, node := range chain {
		width := int(40 * float64(node.Took) / float64(longest))
		fmt.Fprintf(writer, "%-24s %s\n", middle(node.Unit, 24), strings.Repeat("#", width))
	}
}

func bootClock(times parse.Times) time.Duration {
	if times.TargetAfter > 0 {
		return times.Firmware + times.Loader + times.Kernel + times.Initrd + times.TargetAfter
	}
	return times.Total
}

func middle(value string, width int) string {
	if len(value) <= width {
		return value
	}
	return value[:width/2-1] + "…" + value[len(value)-width/2:]
}
