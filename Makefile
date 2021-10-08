CURRENT_REVISION ?= $(shell git rev-parse --short HEAD)
LDFLAGS = -w -s -X 'main.version=Unknown' -X 'main.gitcommit=$(CURRENT_REVISION)'

all: clean test build

test:
	go test ./...

tidy:
	go mod tidy -v

lint:
	golangci-lint run ./...

build:
	go build -ldflags="$(LDFLAGS)" -trimpath -o bin/s3get ./cmd/s3get

clean:
	rm -rf bin dist

.PHONY: test build cross deploy clean
