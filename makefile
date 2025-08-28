.PHONY: all build test lint clean

all: build

build:
	go build -o bin/certinfo ./cmd/certinfo
	go build -o bin/certinfo-web ./cmd/certinfo-web

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/