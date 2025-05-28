package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/anuj/build-analyzer/app/build-analyzer/analyzer"
	"github.com/anuj/build-analyzer/app/build-analyzer/cache"
	"github.com/anuj/build-analyzer/app/build-analyzer/jenkins"

	"github.com/gin-gonic/gin"
)

var (
	jenkinsClient *jenkins.Client
	buildCache    *cache.Cache
)

func main() {
	// Add client mode flags
	clientMode := flag.Bool("client", false, "Run in client mode")
	buildID := flag.String("build", "", "Build ID to analyze")
	serverURL := flag.String("server", "http://localhost:8080", "Log server URL")
	flag.Parse()

	if *clientMode {
		runClient(*buildID, *serverURL)
	} else {
		runServer()
	}
}

func runServer() {
	os.MkdirAll("build-logs", 0755)
	jenkinsClient = jenkins.NewClient()
	buildCache = cache.NewCache()

	r := gin.Default()
	r.HEAD("/logs/:buildId", downloadBuildLog)
	r.GET("/logs/:buildId", analyzeBuildLog)
	r.Run(":8080")
}

func runClient(buildID, serverURL string) {
	// List available log files
	files, err := os.ReadDir("build-logs")

	if len(files) == 0 {
		fmt.Println("No log files available. Download some logs first.")
		os.Exit(1)
	}

	if buildID == "" {
		fmt.Println("Available log files:")
		for i, file := range files {
			fmt.Printf("[%d] %s\n", i+1, strings.TrimSuffix(file.Name(), ".log"))
		}

		var choice int
		fmt.Print("\nSelect a log file (1-", len(files), "): ")
		fmt.Scan(&choice)

		if choice < 1 || choice > len(files) {
			fmt.Println("Invalid selection")
			os.Exit(1)
		}

		buildID = strings.TrimSuffix(files[choice-1].Name(), ".log")
	}

	// Download and analyze logs

	url := fmt.Sprintf("%s/logs/%s", serverURL, buildID)
	fmt.Printf("Analyzing build ID: %s\n", buildID)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading logs: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var analysis map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&analysis); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	printAnalysis(buildID, analysis)
}

func printAnalysis(buildID string, analysis map[string]interface{}) {
	fmt.Printf("\nBuild Analysis for %s:\n", buildID)
	fmt.Printf("Status: %s\n", analysis["status"])
	fmt.Printf("Duration: %s\n", analysis["duration"])
	fmt.Printf("Content Length: %v bytes\n", analysis["contentLength"])
	fmt.Printf("\nError Summary:\n")
	fmt.Printf("Total Errors: %v\n", analysis["totalErrors"])
	fmt.Printf("Unique Errors: %v\n", analysis["uniqueErrors"])
	fmt.Printf("\nWarning Summary:\n")
	fmt.Printf("Total Warnings: %v\n", analysis["totalWarnings"])
	fmt.Printf("Unique Warnings: %v\n", analysis["uniqueWarnings"])

	fmt.Printf("\nTop Common Errors:\n")
	if commonErrors, ok := analysis["commonErrors"].([]interface{}); ok {
		for _, err := range commonErrors {
			if e, ok := err.(map[string]interface{}); ok {
				fmt.Printf("[%v occurrences] %s\n", e["count"], e["message"])
			}
		}
	}
}

func downloadBuildLog(c *gin.Context) {
	buildId := c.Param("buildId")

	resp, err := jenkinsClient.HeadBuildLog(buildId)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()

	// Copy relevant headers from Jenkins response
	c.Header("Content-Length", fmt.Sprintf("%d", resp.ContentLength))
	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Header("Last-Modified", resp.Header.Get("Last-Modified"))

	c.Status(http.StatusOK)
}

func analyzeBuildLog(c *gin.Context) {
	buildId := c.Param("buildId")

	// Check cache first
	if analysis, found := buildCache.Get(buildId); found {
		c.JSON(http.StatusOK, analysis)
		return
	}

	// Download if not exists
	logPath, err := jenkinsClient.DownloadBuildLog(buildId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Analyze log
	analysis, err := analyzer.AnalyzeBuildLog(logPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Cache results
	buildCache.Set(buildId, analysis)

	c.JSON(http.StatusOK, analysis)
}
