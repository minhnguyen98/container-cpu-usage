#!/bin/bash

set -e
set -x

GO_FILES=$(shell \
	find . '(' -path '*/.*' -o -path './vendor' ')' -prune \
	-o -name '*.go' -print | cut -b3-)

# Log Go configuration.
go version
go env
go mod download

if [ ! -x "$(pwd)/bin/staticcheck" ]
then
  GOBIN=$(pwd)/bin go install honnef.co/go/tools/cmd/staticcheck@latest
fi

if [ ! -x "$(pwd)/bin/golangci-lint" ]; then
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(pwd)/bin" v1.56.2
fi

rm -rf lint.log
echo "Checking gofmt"
gofmt -d -s "$GO_FILES" 2>&1 | tee lint.log
echo "Checking go vet"
go vet ./... 2>&1 | tee -a lint.log
echo "Checking golint"
"$(pwd)"/bin/golangci-lint run | tee -a lint.log
echo "Checking staticcheck"
"$(pwd)"/bin/staticcheck ./... 2>&1 |  tee -a lint.log
