.PHONY: go-build go-test go-lint go-fmt go-check distclean

GO_PACKAGES := $(shell go list ./...)
GIT_COMMIT ?= $(shell { git stash create; git rev-parse HEAD; } | grep -Exm1 '[[:xdigit:]]{40}')
export SOURCE_DATE_EPOCH ?= $(shell git show -s --format="%ct" $(GIT_COMMIT))
VERSION ?= $(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match)
VERSION_PKG = github.com/aybykovskii/gitlab-tui/internal/version
export LDFLAGS += -X $(VERSION_PKG).GitCommit=$(GIT_COMMIT)
export LDFLAGS += -X $(VERSION_PKG).SourceDateEpoch=$(SOURCE_DATE_EPOCH)
export LDFLAGS += -X $(VERSION_PKG).Version=$(VERSION)
export LDFLAGS += -s -w
export CGO_ENABLED ?= 0

build:
	go build -ldflags='$(LDFLAGS)' ./cmd/gitlab-tui

test:
	@go clean -testcache
	CGO_ENABLED=1 go test -race $(GO_PACKAGES)

fmt:
	gofumpt -w cmd internal
	golangci-lint run --fix ./...
	gofumpt -w cmd internal

lint:
	@test -z "$$(gofumpt -l cmd internal)" || (gofumpt -l cmd internal && exit 1)
	golangci-lint run ./...

check: lint test build

distclean:
	go clean -x -cache -testcache -modcache ./...
