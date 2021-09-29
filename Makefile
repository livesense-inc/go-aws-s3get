REPO = go-aws-s3get
OWNER = etsxxx
BIN = s3get
CURRENT_REVISION ?= $(shell git rev-parse --short HEAD)
LDFLAGS = -w -s -X 'main.version=Unknown' -X 'main.gitcommit=$(CURRENT_REVISION)'

all: clean test build

test:
	go test ./...

build:
	go build -ldflags="$(LDFLAGS)" -trimpath -o bin/$(BIN) ./cmd/s3get

clean:
	rm -rf bin dist

.PHONY: test build cross deploy clean
