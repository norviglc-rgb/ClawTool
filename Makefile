GO ?= go

.PHONY: fmt lint test test-unit test-integration test-e2e test-i18n build build-all generate check release-dry-run

fmt:
	$(GO) fmt ./...

lint:
	$(GO) vet ./...

test: test-unit test-integration test-i18n

test-unit:
	$(GO) test ./...

test-integration:
	$(GO) test ./...

test-e2e:
	$(GO) test ./...

test-i18n:
	$(GO) test ./internal/i18n/... ./internal/schema/...

build:
	$(GO) build ./cmd/clawtool

build-all:
	$(GO) build ./...

generate:
	$(GO) test ./...

check: fmt lint test

release-dry-run: build-all

