NAME="github.com/raystack/guardian"
COMMIT := $(shell git rev-parse --short HEAD)
TAG := "$(shell git rev-list --tags --max-count=1)"
VERSION := "$(shell git describe --tags ${TAG})-next"
BUILD_DIR=dist
PROTON_COMMIT := "d382bec8545902c081435a931c346a8543583aaf"

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

clean: tidy ## Clean the build artifacts
	@echo "Cleaning up build directories..."
	@rm -rf $coverage.out ${BUILD_DIR}

test:  ## Run the tests
	go test ./... -race -coverprofile=coverage.out

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
	go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1
	go get github.com/golang/protobuf/proto@v1.5.2
	go get github.com/golang/protobuf/protoc-gen-go@v1.5.2
	go get google.golang.org/grpc@v1.40.0
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.5.0
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.5.0
	go get github.com/bufbuild/buf/cmd/buf@v1.15.1

help: ## Display this help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
