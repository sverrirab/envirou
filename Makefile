all: build test

build:
	go build .

test:
	go test ./...

install:
	go install .
