.PHONY: run build test tidy lint seed

GO ?= go

run:
	$(GO) run ./cmd/server

seed:
	$(GO) run ./cmd/seed

build:
	$(GO) build -o bin/metrics-server ./cmd/server
	$(GO) build -o bin/metrics-seed ./cmd/seed

test:
	$(GO) test ./...

tidy:
	$(GO) mod tidy

lint:
	$(GO) vet ./...
