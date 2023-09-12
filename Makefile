NAME="github.com/goto/guardian"
COMMIT := $(shell git rev-parse --short HEAD)
TAG := "$(shell git rev-list --tags --max-count=1)"
VERSION := "$(shell git describe --tags ${TAG})-next"
BUILD_DIR=dist
PROTON_COMMIT := "8300157e4ac12a15f0dcc7f02b4b80c3c76be9fe"

.PHONY: all build clean test tidy vet proto setup format generate

all: clean test build format lint

tidy:
	@echo "Tidy up go.mod..."
	@go mod tidy -v

install:
	@echo "Installing Guardian to ${GOBIN}..."
	@go install

format:
	@echo "Running go fmt..."
	@go fmt

lint:
	@echo "Running lint checks using golangci-lint..."
	@golangci-lint run

clean: tidy
	@echo "Cleaning up build directories..."
	@rm -rf $coverage.out ${BUILD_DIR}

test:
	go test ./... -race -coverprofile=coverage.out

test-short:
	@echo "Running short tests by disabling store tests..."
	go test ./... -race -short -coverprofile=coverage.out

coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out

build:
	@echo "Building guardian version ${VERSION}..."
	go build -ldflags "-X ${NAME}/core.Version=${VERSION} -X ${NAME}/core.BuildCommit=${COMMIT}" -o dist/guardian .
	@echo "Build complete"

buildr: setup
	goreleaser --snapshot --skip-publish --rm-dist

vet:
	go vet ./...

download:
	@go mod download

generate:
	@echo "Running go generate..."
	go generate ./...

config:
	@echo "Initializing sample server config..."
	@cp internal/server/config.yaml config.yaml

proto:
	@echo "Generating protobuf from goto/proton"
	@echo " [info] make sure correct version of dependencies are installed using 'make install'"
	@buf generate https://github.com/goto/proton/archive/${PROTON_COMMIT}.zip#strip_components=1 --template buf.gen.yaml --path gotocompany/guardian
	@echo "Protobuf compilation finished"

setup:
	@echo "Installing dependencies..."
	go mod tidy
	go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1
	go get github.com/golang/protobuf/proto@v1.5.2
	go get github.com/golang/protobuf/protoc-gen-go@v1.5.2
	go get google.golang.org/grpc@v1.40.0
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.5.0
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.5.0
	go get github.com/bufbuild/buf/cmd/buf@v1.15.1
