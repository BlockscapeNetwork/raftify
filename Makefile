# Unit Tests
tests:
	@echo "Running unit tests for Raftify..."
	@go test -v -cover -coverprofile=coverage.txt -covermode=atomic -short ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Tests finished"