export GOBIN ?= $(shell pwd)/bin

.PHONY: build
build:
	go build ./...

.PHONY: install
install:
	go mod download

.PHONY: test
test:
	go test -v ./...

.PHONY: cover
cover:
	go test -coverprofile=cover.out -covermode=atomic -coverpkg=./... ./...
	go tool cover -html=cover.out -o cover.html

.PHONY: lint
lint: 
	./scripts/lint.sh
