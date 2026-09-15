.PHONY: run build test fmt fmt-check vet check

run:
	go run ./cmd/api

build:
	go build ./...

test:
	go test -race -shuffle=on ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

fmt-check:
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))" || \
		(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'); exit 1)

vet:
	go vet ./...

check: fmt-check vet test build
