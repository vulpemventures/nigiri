.PHONY: install build release dry-release clean cov fmt help vet test

## install: installs dependencies
install:
	export GO111MODULE=on
	chmod u+x ./scripts/install
	./scripts/install	

## build: build binary for ARM
build:
	export GO111MODULE=on
	chmod u+x ./scripts/build
	./scripts/build

release:
	goreleaser

dry-release:
	goreleaser --snapshot --skip-publish --rm-dist

## clean: cleans the binary
clean:
	@echo "Cleaning..."
	export GO111MODULE=on
	chmod u+x ./scripts/clean
	./scripts/clean

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

## test-ci: runs travis tests
test-ci:
	@echo "Testing..."
	go test -short -v ./...

## cov: generates coverage report
cov:
	@echo "Coverage..."
	go test -cover ./...
