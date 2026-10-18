package parse

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"
)

var ErrBootInProgress = errors.New("boot is not finished")

type Times struct {
	Firmware, Loader, Kernel, Initrd, Userspace, Total time.Duration
	Target                                             string
	TargetAfter                                        time.Duration
}

var phasePattern = regexp.MustCompile(`([0-9.]+(?:us|ms|s|min|h|d)(?:\s*[0-9.]+(?:us|ms|s|min|h|d))*) \((firmware|loader|kernel|initrd|userspace)\)`)
var totalPattern = regexp.MustCompile(`=\s*([0-9.]+(?:us|ms|s|min|h|d)(?:\s*[0-9.]+(?:us|ms|s|min|h|d))*)`)
var targetPattern = regexp.MustCompile(`^(.+?) reached after (.+?) in userspace\.?$`)

// ParseTime parses systemd-analyze time output. Missing phases remain zero.
func ParseTime(reader io.Reader) (Times, error) {
	var result Times
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimSuffix(scanner.Text(), "\r"))
		if strings.Contains(strings.ToLower(line), "bootup is not yet finished") {
			return Times{}, ErrBootInProgress
		}
		for _, match := range phasePattern.FindAllStringSubmatch(line, -1) {
			span, err := Span(match[1])
			if err != nil {
				continue
			}
			switch match[2] {
			case "firmware":
				result.Firmware = span
			case "loader":
				result.Loader = span
			case "kernel":
				result.Kernel = span
			case "initrd":
				result.Initrd = span
			case "userspace":
				result.Userspace = span
			}
		}
		if match := totalPattern.FindStringSubmatch(line); match != nil {
			if span, err := Span(match[1]); err == nil {
				result.Total = span
			}
		}
		if match := targetPattern.FindStringSubmatch(line); match != nil {
			if span, err := Span(match[2]); err == nil {
				result.Target, result.TargetAfter = match[1], span
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return Times{}, err
	}
	return result, nil
}
