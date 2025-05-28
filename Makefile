.PHONY: build run-server run-client test clean

build:
	go build -o bin/build-analyzer app/build-analyzer/main.go

run-server:
	./bin/build-analyzer

run-client:
	./bin/build-analyzer -client

test:
	go test ./tests/main_test.go
	go test ./tests/...

test-coverage:
	go test ./tests/... -coverprofile=coverage.out
	go tool cover -html=coverage.out

clean:
	rm -rf bin/
	rm -rf build-logs/
	mkdir -p build-logs/
