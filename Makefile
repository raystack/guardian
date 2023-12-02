NAME="github.com/raystack/guardian"
COMMIT := $(shell git rev-parse --short HEAD)
TAG := "$(shell git rev-list --tags --max-count=1)"
VERSION := "$(shell git describe --tags ${TAG})-next"
BUILD_DIR=dist
PROTON_COMMIT := "bd2a1d201fb4931e7b62d93031cb541016818daa"

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

lint: ## Lint checker
	@echo "Running lint checks using golangci-lint..."
	@golangci-lint run

lintf: ## Lint checker and fix
	@echo "Running lint checks using golangci-lint..."
	@golangci-lint run --fix

clean: tidy ## Clean the build artifacts
	@echo "Cleaning up build directories..."
	@rm -rf $coverage.out ${BUILD_DIR}

test: tidy ## Run the tests
	go test ./... -race -coverprofile=coverage.out

test-short:
	@echo "Running short tests by disabling store tests..."
	go test ./... -race -short -coverprofile=coverage.out

coverage: test ## Print the code coverage
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out

build: ## Build the guardian binary
	@echo "Building guardian version ${VERSION}..."
	go build -ldflags "-X ${NAME}/core.Version=${VERSION} -X ${NAME}/core.BuildCommit=${COMMIT}" -o dist/guardian .
	@echo "Build complete"

buildr: setup
	goreleaser --snapshot --skip-publish --rm-dist

vet:
	go vet ./...

download:
	@echo Download go.mod dependencies
	@go mod download

generate: ## Run all go generate in the code base
	@echo "Running go generate..."
	go generate ./...

config: ## Generate the sample config file
	@echo "Initializing sample server config..."
	@cp internal/server/config.yaml config.yaml

proto: ## Generate the protobuf files
	@echo "Generating protobuf from raystack/proton"
	@echo " [info] make sure correct version of dependencies are installed using 'make install'"
	@buf generate https://github.com/raystack/proton/archive/${PROTON_COMMIT}.zip#strip_components=1 --template buf.gen.yaml --path raystack/guardian
	@echo "Protobuf compilation finished"

setup: ## Install all the dependencies
	@echo "Installing dependencies..."
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
	go install github.com/golang/protobuf/proto@v1.5.2
	go install github.com/golang/protobuf/protoc-gen-go@v1.5.2
	go install google.golang.org/grpc@v1.40.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.29.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.5.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.5.0
	go install github.com/bufbuild/buf/cmd/buf@v1.29.0
	go install github.com/vektra/mockery/v2@v2.38.0

help: ## Display this help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

doc: ## Generate api documentation
	@echo ">genetate api docs"
	@cd $(CURDIR)/docs/docs; yarn docusaurus clean-api-docs all;  yarn docusaurus gen-api-docs all

doc-build: ## Run documentation locally
	@echo "> building docs"
	@cd $(CURDIR)/docs/docs; yarn start