.PHONY: go-build go-test go-lint go-fmt go-check

GO_PACKAGES := $(shell go list ./... | grep -v '/node_modules/')

go-build:
	go build ./cmd/gitlab-tui-go

go-test:
	go test $(GO_PACKAGES)

go-fmt:
	gofumpt -w cmd internal

go-lint:
	@test -z "$$(gofumpt -l cmd internal)" || (gofumpt -l cmd internal && exit 1)
	go vet $(GO_PACKAGES)
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run $(GO_PACKAGES); else echo "golangci-lint not installed; skipped"; fi

go-check: go-lint go-test go-build
