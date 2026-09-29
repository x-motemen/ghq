CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X main.revision=$(CURRENT_REVISION)"
VERBOSE_FLAG = $(if $(VERBOSE),-v)

.PHONY: deps
deps:
	go mod download

.PHONY: credits-deps
credits-deps:
	go install github.com/Songmu/gocredits/cmd/gocredits@v1.0.1

.PHONY: devel-deps
devel-deps: credits-deps
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
prepare-release: credits-deps
	go mod tidy
	gocredits -w
	git update-index --add --remove -- go.mod go.sum CREDITS

CREDITS: go.sum credits-deps
	gocredits -w
