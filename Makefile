NAME="github.com/odpf/guardian"
LAST_COMMIT := $(shell git rev-parse --short HEAD)
LAST_TAG := "$(shell git rev-list --tags --max-count=1)"
APP_VERSION := "$(shell git describe --tags ${LAST_TAG})-next"
PROTON_COMMIT := "bb41ca0a800f039208f21730ab9e78b49bf08aa7"

.PHONY: all build test clean dist vet proto install

all: build

build: ## Build the guardian binary
	@echo " > building guardian version ${APP_VERSION}"
	go build -ldflags "-X ${NAME}/app.Version=${APP_VERSION} -X ${NAME}/app.BuildCommit=${LAST_COMMIT}" -o guardian .
	@echo " - build complete"

buildr: install ## Build with goreleaser
	goreleaser --snapshot --skip-publish --rm-dist

test: ## Run the tests
	go test ./... -race -coverprofile=coverage.out

coverage: ## Print code coverage
	go test -race -coverprofile coverage.txt -covermode=atomic ./... & go tool cover -html=coverage.out

vet: ## Run the go vet tool
	go vet ./...

lint: ## Lint with golangci-lint
	golangci-lint run

generate: ## Generate mocks
	go generate ./...

proto: ## Generate the protobuf files
	@echo " > generating protobuf from odpf/proton"
	@echo " > [info] make sure correct version of dependencies are installed using 'make install'"
	@buf generate https://github.com/odpf/proton/archive/${PROTON_COMMIT}.zip#strip_components=1 --template buf.gen.yaml --path odpf/guardian
	@echo " > protobuf compilation finished"

clean: ## Clean the build artifacts
	rm -rf guardian dist/
	rm -f coverage.*

install: ## install required dependencies
	@echo "> installing dependencies"
	go mod tidy
	go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1
	go get github.com/golang/protobuf/proto@v1.5.2
	go get github.com/golang/protobuf/protoc-gen-go@v1.5.2
	go get google.golang.org/grpc@v1.40.0
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.5.0
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.5.0
	go get github.com/bufbuild/buf/cmd/buf@v0.54.1

help: ## Display this help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
