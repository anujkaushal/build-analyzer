package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/anuj/build-analyzer/app/build-analyzer/analyzer"
	"github.com/anuj/build-analyzer/app/build-analyzer/cache"
	"github.com/anuj/build-analyzer/app/build-analyzer/jenkins"
	"github.com/gin-gonic/gin"
)

var (
	testJenkinsClient *jenkins.Client
	testBuildCache    *cache.Cache
)

func TestMain(m *testing.M) {
	// Setup
	gin.SetMode(gin.TestMode)
	os.MkdirAll("build-logs", 0755)

	// Initialize test dependencies
	testJenkinsClient = jenkins.NewClient()
	testBuildCache = cache.NewCache()

	// Run tests
	code := m.Run()

	// Cleanup
	os.RemoveAll("build-logs")
	os.Exit(code)
}

func setupTestRouter() *gin.Engine {
	r := gin.Default()

	// Use actual handlers with test dependencies
	r.HEAD("/logs/:buildId", func(c *gin.Context) {
		buildId := c.Param("buildId")
		resp, err := testJenkinsClient.HeadBuildLog(buildId)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer resp.Body.Close()

		c.Header("Content-Length", resp.Header.Get("Content-Length"))
		c.Header("Content-Type", resp.Header.Get("Content-Type"))
		c.Header("Last-Modified", resp.Header.Get("Last-Modified"))
		c.Status(http.StatusOK)
	})

	r.GET("/logs/:buildId", func(c *gin.Context) {
		buildId := c.Param("buildId")

		// Check cache
		if analysis, found := testBuildCache.Get(buildId); found {
			c.JSON(http.StatusOK, analysis)
			return
		}

		// Create test log file
		logPath := "build-logs/test.log"
		os.WriteFile(logPath, []byte("test log content\nERROR: test error"), 0644)

		// Analyze log
		analysis, err := analyzer.AnalyzeBuildLog(logPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		testBuildCache.Set(buildId, analysis)
		c.JSON(http.StatusOK, analysis)
	})

	return r
}

func TestEndpoints(t *testing.T) {
	router := setupTestRouter()

	t.Run("HEAD request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("HEAD", "/logs/test-build", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if len(w.Header().Get("Content-Length")) == 0 {
			t.Error("Missing Content-Length header")
		}
	})

	t.Run("GET request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/logs/test-build", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response analyzer.BuildAnalysis
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}
	})
}
