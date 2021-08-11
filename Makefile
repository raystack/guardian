NAME="github.com/odpf/guardian"
LAST_COMMIT := $(shell git rev-parse --short HEAD)
LAST_TAG := "$(shell git rev-list --tags --max-count=1)"
OPMS_VERSION := "$(shell git describe --tags ${LAST_TAG})-next"
COVERFILE="/tmp/guardian.coverprofile"

.PHONY: all build test clean

all: build

build:
	go build -ldflags "-X ${NAME}/config.Version=${OPMS_VERSION} -X ${NAME}/config.BuildCommit=${LAST_COMMIT}" -o guardian .

clean:
	rm -rf guardian dist/

test:
	go test ./... -coverprofile=coverage.out

test-coverage: test
	go tool cover -html=coverage.out

dist:
	@bash ./scripts/build.sh

generate-proto: ## regenerate protos
	@echo " > cloning protobuf from odpf/proton"
	@rm -rf proton/
	@git -c advice.detachedHead=false clone https://github.com/odpf/proton --depth 1 --quiet --branch main
	@echo " > generating protobuf"
	@echo " > info: make sure correct version of dependencies are installed using 'install'"
	@buf generate
	@echo " > protobuf compilation finished"
