package jenkins

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Make it public for testing
var JenkinsBaseURL = "https://ci.jenkins.io"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

func (c *Client) HeadBuildLog(buildID string) (*http.Response, error) {
	// Download the log file first
	logPath, err := c.DownloadBuildLog(buildID)
	if err != nil {
		return nil, err
	}

	// Get file stats
	fileInfo, err := os.Stat(logPath)
	if err != nil {
		return nil, err
	}

	// Create a dummy response with the file information
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       http.NoBody,
	}

	resp.Header.Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	resp.Header.Set("Content-Type", "text/plain")
	resp.Header.Set("Last-Modified", fileInfo.ModTime().UTC().Format(http.TimeFormat))
	resp.ContentLength = fileInfo.Size()

	return resp, nil
}

func (c *Client) DownloadBuildLog(buildID string) (string, error) {
	fileName := filepath.Join("build-logs", fmt.Sprintf("%s.log", buildID))

	// Check if file already exists
	if _, err := os.Stat(fileName); err == nil {
		fmt.Println("Using cached build log:", fileName)
		return fileName, nil
	}

	url := fmt.Sprintf("%s/job/Tools/job/bom/job/PR-5113/%s/consoleText", JenkinsBaseURL, buildID)
	fmt.Println("Downloading build log from:", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return fileName, err
}
