package analyzer_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/anuj/build-analyzer/app/build-analyzer/analyzer"
)

func TestAnalyzeBuildLog(t *testing.T) {
	// Create temp test log file
	content := `2023-11-15 10:00:01 ERROR Failed to compile
2023-11-15 10:00:02 ERROR Failed to compile
2023-11-15 10:00:03 WARNING Package not found
Finished: SUCCESS
Total time: 5 minutes`

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	analysis, err := analyzer.AnalyzeBuildLog(logFile)
	fmt.Println("Analysis Result:")
	fmt.Println(analysis)
	fmt.Println("End Here.")
	if err != nil {
		t.Fatal(err)
	}

	// Test full analysis results
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Status", analysis.Status, "SUCCESS"},
		{"Duration", analysis.Duration, "5 minutes"},
		{"TotalErrors", analysis.TotalErrors, 2},
		{"UniqueErrors", analysis.UniqueErrors, 1},
		{"TotalWarnings", analysis.TotalWarnings, 1},
		{"UniqueWarnings", analysis.UniqueWarnings, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.got)
			}
		})
	}

	// Test common errors
	if len(analysis.CommonErrors) != 1 {
		t.Errorf("Expected 1 common error, got %d", len(analysis.CommonErrors))
	}

	if len(analysis.CommonErrors) > 0 {
		commonError := analysis.CommonErrors[0]
		if commonError.Count != 2 {
			t.Errorf("Expected error count 2, got %d", commonError.Count)
		}
		if commonError.FirstSeen != "line 1" {
			t.Errorf("Expected first seen at line 1, got %s", commonError.FirstSeen)
		}
	}
}
