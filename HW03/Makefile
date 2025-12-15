.PHONY: build
build:
	go build -o bin/reelgoofy cmd/reelgoofy/main.go

.PHONY: test
test:
	go test -cover -race -timeout 30s ./...

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint run

.PHONY: format
format:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint fmt

.PHONY: format-diff
format-diff:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint fmt --diff-colored
