package jenkins_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/anuj/build-analyzer/app/build-analyzer/jenkins"
)

func TestDownloadBuildLog(t *testing.T) {
	// Create test server with realistic Jenkins response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/consoleText") {
			t.Errorf("Unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("2023-11-15 10:00:01 ERROR Test error\nFinished: SUCCESS\nTotal time: 1min"))
	}))
	defer server.Close()

	// Override Jenkins base URL for testing
	oldURL := jenkins.JenkinsBaseURL
	jenkins.JenkinsBaseURL = server.URL
	defer func() { jenkins.JenkinsBaseURL = oldURL }()

	client := jenkins.NewClient()

	// Clean up test files
	testDir := "build-logs"
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	// Test first download
	path, err := client.DownloadBuildLog("test-build")
	if err != nil {
		t.Fatal(err)
	}

	// Verify content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "ERROR Test error") {
		t.Errorf("Expected error message in content, got %q", string(content))
	}

	// Test cached download
	path2, err := client.DownloadBuildLog("test-build")
	if err != nil {
		t.Fatal(err)
	}
	if path != path2 {
		t.Error("Expected same path for cached download")
	}
}
