package parse

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"time"
)

type ChainNode struct {
	Unit     string
	At, Took time.Duration
	Depth    int
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*[[:alpha:]]`)
var chainPattern = regexp.MustCompile(`^([ │├└─` + "`" + `\+\|\-\\]*)(\S+?)(?:\s+@(.+?))?(?:\s+\+(.+))?$`)

// ParseChain parses systemd-analyze critical-chain output.
func ParseChain(reader io.Reader) ([]ChainNode, error) {
	var nodes []ChainNode
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimRight(strings.TrimSuffix(ansiPattern.ReplaceAllString(scanner.Text(), ""), "\r"), " \t")
		match := chainPattern.FindStringSubmatch(line)
		if match == nil || !strings.Contains(match[2], ".") {
			continue
		}

		node := ChainNode{Unit: match[2], Depth: chainDepth(match[1])}
		if match[3] != "" {
			if span, err := Span(strings.TrimSpace(match[3])); err == nil {
				node.At = span
			}
		}
		if match[4] != "" {
			if span, err := Span(strings.TrimSpace(match[4])); err == nil {
				node.Took = span
			}
		}
		nodes = append(nodes, node)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nodes, nil
}

func chainDepth(prefix string) int {
	depth := strings.Count(prefix, "├") + strings.Count(prefix, "└") + strings.Count(prefix, "`") + strings.Count(prefix, "+") + strings.Count(prefix, "\\")
	if depth == 0 {
		return 0
	}
	return depth + strings.Count(prefix, "  ")
}
