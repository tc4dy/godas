.PHONY: test bench cover lint build clean install

test:
	go test -race -cover ./...

bench:
	go test -bench=. -benchmem ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run

build:
	go build -o bin/godas ./cmd/godas

install:
	go install ./cmd/godas

clean:
	rm -rf bin/ coverage.out

all: lint test build