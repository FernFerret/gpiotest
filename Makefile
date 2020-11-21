VERSION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GOLDFLAGS += -X main.version=$(VERSION)
GOLDFLAGS += -X main.buildTime=$(BUILDTIME)
GOFLAGS = -ldflags "$(GOLDFLAGS)"

build:
	GOOS=linux GOARCH=arm GOARM=6 go build -o bin/gpiotest $(GOFLAGS) ./cmd/gpiotest

clean:
	rm -f bin/*

all: build
