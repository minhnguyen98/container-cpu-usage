export GOBIN ?= $(shell pwd)/bin

GO_FILES := $(shell \
	find . '(' -path '*/.*' -o -path './vendor' ')' -prune \
	-o -name '*.go' -print | cut -b3-)

GOLINT = $(GOBIN)/golint
STATICCHECK = $(GOBIN)/staticcheck

.PHONY: build
build:
	go build ./...

.PHONY: install
install:
	go mod download

.PHONY: test
test:
	go test -v ./...

$(GOLINT): tools/go.mod
	cd tools && go install golang.org/x/lint/golint

$(STATICCHECK): tools/go.mod
	cd tools && go install honnef.co/go/tools/cmd/staticcheck@2023.1.2

.PHONY: lint
lint: $(GOLINT) $(STATICCHECK)
	@rm -rf lint.log
	@echo "Checking gofmt"
	@gofmt -d -s $(GO_FILES) 2>&1 | tee lint.log
	@echo "Checking go vet"
	@go vet ./... 2>&1 | tee -a lint.log
	@echo "Checking golint"
	@$(GOLINT) ./... | tee -a lint.log
	@echo "Checking staticcheck"
	@$(STATICCHECK) ./... 2>&1 |  tee -a lint.log
	@[ ! -s lint.log ]
