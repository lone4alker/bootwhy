package analyze

import (
	"sort"
	"time"

	"github.com/lone4alker/bootwhy/internal/acquire"
	"github.com/lone4alker/bootwhy/internal/parse"
)

type Shape int

const (
	Normal Shape = iota
	Single
	Several
	SharedBottleneck
	MissingData
)

type Contributor struct {
	Unit                  string
	Took, Expected, Added time.Duration
	Share                 float64
}

type Result struct {
	Shape        Shape
	Contributors []Contributor
	Rest         time.Duration
	Noted        []parse.BlameEntry
}

// Analyze identifies chain units whose duration exceeds comparable peers.
func Analyze(_ parse.Times, chain []parse.ChainNode, blame []parse.BlameEntry, context acquire.Context) Result {
	if len(chain) == 0 {
		return Result{Shape: MissingData}
	}
	durations := make([]time.Duration, 0, len(chain))
	chainUnits := make(map[string]bool, len(chain))
	for _, node := range chain {
		chainUnits[node.Unit] = true
		if node.Took > 0 {
			durations = append(durations, node.Took)
		}
	}
	if len(durations) == 0 {
		return Result{Shape: MissingData}
	}
	expected := max(MinExpected, time.Duration(PeerFactor*float64(median(durations))))
	if context.VM {
		expected = time.Duration(float64(expected) * VMFactor)
	}
	if context.Rotational {
		expected = time.Duration(float64(expected) * HDDFactor)
	}

	var result Result
	for _, node := range chain {
		unitExpected := min(expected, node.Took)
		added := node.Took - unitExpected
		if added > 0 {
			result.Contributors = append(result.Contributors, Contributor{Unit: node.Unit, Took: node.Took, Expected: unitExpected, Added: added})
		}
	}
	for _, entry := range blame {
		if !chainUnits[entry.Unit] && entry.Took > expected {
			result.Noted = append(result.Noted, entry)
		}
	}
	sort.Slice(result.Contributors, func(i, j int) bool { return result.Contributors[i].Added > result.Contributors[j].Added })
	var total time.Duration
	for _, contributor := range result.Contributors {
		total += contributor.Added
	}
	if total == 0 || total < NormalAbs {
		result.Shape = Normal
		return result
	}
	kept := result.Contributors[:0]
	for _, contributor := range result.Contributors {
		share := float64(contributor.Added) / float64(total)
		if contributor.Added < FloorAbs && share < FloorShare {
			result.Rest += contributor.Added
			continue
		}
		contributor.Share = share
		kept = append(kept, contributor)
	}
	result.Contributors = kept
	if len(kept) == 0 {
		result.Shape = Normal
		return result
	}
	if len(kept) >= BottleneckUnits {
		result.Shape = SharedBottleneck
	} else if kept[0].Share >= DominantShare {
		result.Shape = Single
	} else {
		result.Shape = Several
	}
	return result
}

func median(values []time.Duration) time.Duration {
	values = append([]time.Duration(nil), values...)
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	return values[len(values)/2]
}

func min(left, right time.Duration) time.Duration {
	if left < right {
		return left
	}
	return right
}

func max(left, right time.Duration) time.Duration {
	if left > right {
		return left
	}
	return right
}
