.PHONY: build
build:
	go build -o bin/ca-service ./cmd/ca-service

.PHONY: test
test:
	go test -cover -race -v -timeout 30s ./...

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.9.0 run

.PHONY: format
format:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.9.0 fmt

.PHONY: format-diff
format-diff:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.9.0 fmt --diff-colored
