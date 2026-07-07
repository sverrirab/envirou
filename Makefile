VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/sverrirab/envirou/cmd.Version=$(VERSION)
BINARY  := envirou

.PHONY: build install test cover vet lint clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

install:
	go install -ldflags "$(LDFLAGS)" .

test:
	go test -race ./... -count=1

cover:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out | tail -1
	@echo "Run 'go tool cover -html=coverage.out' for a detailed report"

vet:
	go vet ./...

lint: vet
	@echo "Lint passed (go vet)"

clean:
	rm -f $(BINARY) coverage.out

all: vet test build
