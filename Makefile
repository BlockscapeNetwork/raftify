# Testing
# Testing for raftify needs to happen sequentially. This makes sure
# the tests are not run in parallel.
tests:
	@echo "Running tests for Raftify..."
	@go test -v -cover -coverprofile=c.out ./...
	@go tool cover -html=c.out -o coverage.html
	@echo "Tests finished"