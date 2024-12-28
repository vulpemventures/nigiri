.PHONY: install clean build release dry-release cov fmt help vet test

## install: installs dependencies
install:
	go mod download
	go mod tidy

## clean: cleans the binary
clean:
	@echo "Cleaning..."
	go clean

## build: build binary
build:
	chmod u+x ./scripts/build
	./scripts/build

## release: build and upload binaries to Github Releases
release:
	goreleaser

## dry-release: build and test goreleaser
dry-release:
	goreleaser --snapshot --skip-publish --rm-dist

## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## fmt: Go Format
fmt:
	@echo "Gofmt..."
	@if [ -n "$(gofmt -l .)" ]; then echo "Go code is not formatted"; exit 1; fi

## vet: code analysis
vet:
	@echo "Vet..."
	@go vet ./...

## test: runs go unit test with default values
test: clean install
	@echo "Testing..."
	go test -v -count=1 -race ./...

## cov: generates coverage report
cov:
	@echo "Coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
