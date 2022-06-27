.PHONY: test run

test:
	@go test ./... -v -race

run:
	@go run ./cmd