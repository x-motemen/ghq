VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X main.revision=$(CURRENT_REVISION)"
VERBOSE_FLAG = $(if $(VERBOSE),-v)
u := $(if $(update),-u)

export GO111MODULE=on

.PHONY: deps
deps:
	go get ${u} -d $(VERBOSE_FLAG)
	go mod tidy

.PHONY: devel-deps
devel-deps: deps
	sh -c '\
	tmpdir=$$(mktemp -d); \
	cd $$tmpdir; \
	go get ${u} \
	  golang.org/x/lint/golint            \
	  github.com/Songmu/godzil/cmd/godzil \
	  github.com/tcnksm/ghr; \
	rm -rf $$tmpdir'

.PHONY: test
test: deps
	go test $(VERBOSE_FLAG) ./...

.PHONY: lint
lint: devel-deps
	golint -set_exit_status ./...

.PHONY: cover
cover: devel-deps
	goveralls

.PHONY: build
build: deps
	go build $(VERBOSE_FLAG) -ldflags=$(BUILD_LDFLAGS)

.PHONY: install
install: deps
	go install $(VERBOSE_FLAG) -ldflags=$(BUILD_LDFLAGS)

.PHONY: release
release: devel-deps
	godzil release

CREDITS: devel-deps go.sum
	godzil credits -w

DIST_DIR = dist/v$(VERSION)
.PHONY: crossbuild
crossbuild: CREDITS
	rm -rf $(DIST_DIR)
	godzil crossbuild -build-ldflags=$(BUILD_LDFLAGS) \
      -include='misc/zsh/_ghq' -z -d $(DIST_DIR)
	cd $(DIST_DIR) && shasum $$(find * -type f -maxdepth 0) > SHASUMS

.PHONY: upload
upload:
	ghr -body="$$(godzil changelog --latest -F markdown)" v$(VERSION) $(DIST_DIR)
