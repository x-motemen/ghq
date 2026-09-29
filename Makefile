CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X main.revision=$(CURRENT_REVISION)"
VERBOSE_FLAG = $(if $(VERBOSE),-v)
u := $(if $(update),-u)

.PHONY: deps
deps:
	go get ${u}

.PHONY: devel-deps
devel-deps:
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: test
test:
	go test $(VERBOSE_FLAG) ./...

.PHONY: lint
lint: devel-deps
	staticcheck ./...

.PHONY: build
build:
	go build $(VERBOSE_FLAG) -ldflags=$(BUILD_LDFLAGS)

.PHONY: install
install:
	go install $(VERBOSE_FLAG) -ldflags=$(BUILD_LDFLAGS)

.PHONY: prepare-release
prepare-release:
	go mod tidy
	gocredits -w
	git update-index --add --remove -- go.mod go.sum CREDITS
