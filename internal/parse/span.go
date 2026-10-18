package parse

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Span parses systemd duration strings such as "1min 2.345s".
func Span(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty span")
	}

	var total time.Duration
	for len(value) > 0 {
		for len(value) > 0 && unicode.IsSpace(rune(value[0])) {
			value = value[1:]
		}
		if value == "" {
			break
		}

		end := 0
		for end < len(value) && (value[end] == '.' || (value[end] >= '0' && value[end] <= '9')) {
			end++
		}
		if end == 0 {
			return 0, fmt.Errorf("invalid span %q", value)
		}
		number, err := strconv.ParseFloat(value[:end], 64)
		if err != nil {
			return 0, fmt.Errorf("parse span %q: %w", value, err)
		}

		unitEnd := end
		for unitEnd < len(value) && unicode.IsLetter(rune(value[unitEnd])) {
			unitEnd++
		}
		if unitEnd == end {
			return 0, fmt.Errorf("missing unit in span %q", value)
		}
		unit := value[end:unitEnd]
		var scale time.Duration
		switch unit {
		case "us":
			scale = time.Microsecond
		case "ms":
			scale = time.Millisecond
		case "s":
			scale = time.Second
		case "min":
			scale = time.Minute
		case "h":
			scale = time.Hour
		case "d":
			scale = 24 * time.Hour
		default:
			return 0, fmt.Errorf("unknown span unit %q", unit)
		}
		total += time.Duration(number * float64(scale))
		value = value[unitEnd:]
	}
	return total, nil
}
