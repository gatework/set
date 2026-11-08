GO ?= go
export GOWORK := off

.PHONY: check test vet fmt-check fmt fuzz lint

check: fmt-check test vet

test:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

fmt-check:
	@test -z "$$(gofmt -l *.go)" || { gofmt -l *.go; exit 1; }

fmt:
	gofmt -w *.go

fuzz:
	$(GO) test -run '^$$' -fuzz '^FuzzSet$$' -fuzztime=30s .

lint:
	golangci-lint run ./...
