NAME="github.com/odpf/guardian"
VERSION=$(shell git describe --always --tags 2>/dev/null)
COVERFILE="/tmp/guardian.coverprofile"

.PHONY: all build test clean

all: build

build:
	go build -ldflags "-X main.Version=${VERSION}" ${NAME}

clean:
	rm -rf guardian dist/

test:
	go test ./... -coverprofile=coverage.out

test-coverage: test
	go tool cover -html=coverage.out

dist:
	@bash ./scripts/build.sh