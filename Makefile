.PHONY: test run

test:
	@go test ./... -v -race ${ARGS}

run:
	@go run ./cmd
