package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

var timestampRegexForAnalysis = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}`)

type ErrorAnalysis struct {
	Message   string `json:"message"`
	Count     int    `json:"count"`
	FirstSeen string `json:"firstSeen"`
}

type BuildAnalysis struct {
	Status         string          `json:"status"`
	Duration       string          `json:"duration"`
	ContentLength  int64           `json:"contentLength"`
	Errors         []string        `json:"errors"`
	Warnings       []string        `json:"warnings"`
	CommonErrors   []ErrorAnalysis `json:"commonErrors"`
	CommonWarnings []ErrorAnalysis `json:"commonWarnings"`
	TotalErrors    int             `json:"totalErrors"`
	TotalWarnings  int             `json:"totalWarnings"`
	UniqueErrors   int             `json:"uniqueErrors"`
	UniqueWarnings int             `json:"uniqueWarnings"`
}

func AnalyzeBuildLog(filePath string) (*BuildAnalysis, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file stats for content length
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	analysis := &BuildAnalysis{
		Errors:        make([]string, 0),
		Warnings:      make([]string, 0),
		ContentLength: fileInfo.Size(),
	}

	errorFreq := make(map[string]*ErrorAnalysis)
	warningFreq := make(map[string]*ErrorAnalysis)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Deduplicate line before analysis
		dedupedLine := dedupLine(line)
		if dedupedLine == "" {
			continue
		}

		switch {
		case strings.Contains(dedupedLine, "ERROR"):
			analysis.TotalErrors++
			if freq, exists := errorFreq[dedupedLine]; exists {
				freq.Count++
			} else {
				errorFreq[dedupedLine] = &ErrorAnalysis{
					Message:   line, // Keep original line for display
					Count:     1,
					FirstSeen: fmt.Sprintf("line %d", lineNum),
				}
				analysis.Errors = append(analysis.Errors, line)
			}

		case strings.Contains(dedupedLine, "WARNING"):
			analysis.TotalWarnings++
			if freq, exists := warningFreq[dedupedLine]; exists {
				freq.Count++
			} else {
				warningFreq[dedupedLine] = &ErrorAnalysis{
					Message:   line,
					Count:     1,
					FirstSeen: fmt.Sprintf("line %d", lineNum),
				}
				analysis.Warnings = append(analysis.Warnings, line)
			}

		case strings.Contains(line, "Finished:"):
			analysis.Status = strings.TrimSpace(strings.Split(line, "Finished:")[1])
		case strings.Contains(line, "Total time:"):
			analysis.Duration = strings.TrimSpace(strings.Split(line, "Total time:")[1])
		}
	}

	// Process common errors
	analysis.UniqueErrors = len(errorFreq)
	analysis.UniqueWarnings = len(warningFreq)
	analysis.CommonErrors = getTopPatterns(errorFreq)
	analysis.CommonWarnings = getTopPatterns(warningFreq)

	return analysis, scanner.Err()
}

func getTopPatterns(freq map[string]*ErrorAnalysis) []ErrorAnalysis {
	patterns := make([]ErrorAnalysis, 0, len(freq))
	for _, v := range freq {
		patterns = append(patterns, *v)
	}

	// Sort by frequency, then by first occurrence
	sort.Slice(patterns, func(i, j int) bool {
		if patterns[i].Count == patterns[j].Count {
			return patterns[i].FirstSeen < patterns[j].FirstSeen
		}
		return patterns[i].Count > patterns[j].Count
	})

	// Return top 10 or all if less than 10
	if len(patterns) > 10 {
		return patterns[:10]
	}
	return patterns
}

// dedupLine removes duplicate spaces and trims the line
func dedupLine(line string) string {
	// Remove timestamp
	line = timestampRegexForAnalysis.ReplaceAllString(line, "")

	// Replace multiple spaces with a single space
	line = strings.Join(strings.Fields(line), " ")
	// Trim spaces from start and end of the line
	return strings.TrimSpace(line)
}
