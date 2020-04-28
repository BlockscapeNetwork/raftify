# Testing
# Testing for raftify needs to happen sequentially. This makes sure
# the tests are not run in parallel.
tests:
	@echo "Running tests for Raftify..."
	@go test -v -coverprofile=/var/folders/65/h72m4ghn4x95qxm_46nbfyhc0000gn/T/vscode-goCW9kIt/go-code-cover ./...
	@echo "Tests finished"