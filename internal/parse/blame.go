package parse

import (
	"bufio"
	"io"
	"strings"
	"time"
)

type BlameEntry struct {
	Unit string
	Took time.Duration
}

// ParseBlame parses systemd-analyze blame output.
func ParseBlame(reader io.Reader) ([]BlameEntry, error) {
	var entries []BlameEntry
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fields := strings.Fields(strings.TrimSuffix(scanner.Text(), "\r"))
		if len(fields) < 2 {
			continue
		}
		lastSplit := 0
		var took time.Duration
		for split := 1; split < len(fields); split++ {
			span, err := Span(strings.Join(fields[:split], " "))
			if err != nil {
				break
			}
			lastSplit, took = split, span
		}
		if lastSplit > 0 {
			entries = append(entries, BlameEntry{Unit: strings.Join(fields[lastSplit:], " "), Took: took})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
