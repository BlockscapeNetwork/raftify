# Unit Tests
unit-tests:
	@echo "Running unit tests for Raftify..."
	@go test -v -cover -coverprofile=coverage.txt -covermode=atomic -short ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Tests finished"

# Integration Tests
integration-tests:
	@echo "Running integration tests for Raftify..."
	@go test -v -parallel=1 helpers_test.go api.go bootstrap.go candidate.go config.go follower.go handlers.go leader.go lists.go messages.go node.go precandidate.go shutdown.go state.go types.go util.go node_integration_test.go
	@echo "Tests finished"