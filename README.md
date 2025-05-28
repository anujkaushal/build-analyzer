# Build Log Analyzer

A Go application that serves as a proxy to download and analyze Jenkins build logs, providing deduplication and analysis features.

## Purpose

- Download and cache Jenkins build logs
- Analyze build logs for errors and warnings
- Deduplicate log entries by removing timestamps and redundant messages
- Provide both HTTP API and console client interfaces
- Cache analysis results to improve performance

## Features

- HTTP API endpoints:
  - `HEAD /logs/{build_id}` - Get log metadata
  - `GET /logs/{build_id}` - Get analyzed log content
- Console client for local log analysis
- Timestamp deduplication
- Error and warning aggregation
- Build status and duration tracking
- Response caching

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd build-analyzer

# Build the application
make build
```

## Usage

### Server Mode
```bash
# Start the server
make run-server
```

### Client Mode
```bash
# Run client with specific build ID
make run-client -build=PR-5113

# Run client interactively
make run-client
```

### Example Client Output
```
Build Analysis for PR-5113:
Status: SUCCESS
Duration: 5 minutes
Content Length: 12345 bytes

Error Summary:
Total Errors: 10
Unique Errors: 3

Warning Summary:
Total Warnings: 5
Unique Warnings: 2

Top Common Errors:
[5 occurrences] Failed to compile module
[3 occurrences] Connection timeout
[2 occurrences] Invalid configuration
```

### Running Tests

```bash
# Run all tests
make test

# Sample test output:
=== RUN   TestAnalyzeBuildLog
    parser_test.go:45: Analysis of build log successful
    parser_test.go:52: Found 2 total errors, 1 unique
=== RUN   TestDownloadBuildLog
    client_test.go:38: Successfully downloaded and cached log
=== RUN   TestEndpoints
    main_test.go:85: All endpoints responding correctly
PASS
coverage: 85.2% of statements
```

## Architecture

- `analyzer/`: Log parsing and analysis logic
- `jenkins/`: Jenkins API client implementation
- `cache/`: In-memory caching system
- `tests/`: Test suite and test utilities

## Error Handling

The application handles various error cases:
- Invalid build IDs
- Network failures
- Invalid log formats
- File system errors

## API Response Format

```json
{
  "status": "SUCCESS",
  "duration": "5 minutes",
  "contentLength": 12345,
  "totalErrors": 10,
  "uniqueErrors": 3,
  "commonErrors": [
    {
      "message": "Failed to compile module",
      "count": 5,
      "firstSeen": "line 42"
    }
  ]
}
```
