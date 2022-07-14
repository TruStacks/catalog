.PHONY: test run

test:
	@go test ./... -v -race ${ARGS}
	@golangci-lint run 

run:
	@go run ./cmd
