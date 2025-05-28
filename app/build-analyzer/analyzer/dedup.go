package analyzer

import (
	"regexp"
	"strings"
)

var (
	timestampRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}`)
	ansiRegex      = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
)

func dedupLineWithRegex(line string) string {
	// Remove timestamp
	line = timestampRegex.ReplaceAllString(line, "")

	// Remove ANSI color codes
	line = ansiRegex.ReplaceAllString(line, "")

	// Remove leading/trailing whitespace
	line = strings.TrimSpace(line)

	return line
}
