.PHONY: go-build go-test go-lint go-fmt go-check

GO_PACKAGES := $(shell go list ./... | grep -v '/node_modules/')

go-build:
	go build ./cmd/gitlab-tui

go-test:
	go test $(GO_PACKAGES)

go-fmt:
	gofumpt -w cmd internal
	golangci-lint run --fix ./...
	gofumpt -w cmd internal

go-lint:
	@test -z "$$(gofumpt -l cmd internal)" || (gofumpt -l cmd internal && exit 1)
	golangci-lint run ./...

go-check: go-lint go-test go-build
