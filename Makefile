# Testing
# Testing for raftify needs to happen sequentially. This makes sure
# the tests are not run in parallel.
tests:
	@echo "Running tests for Raftify..."
	@go test -parallel=1 -v -cover -coverprofile=coverage.txt -covermode=atomic -short ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Tests finished"