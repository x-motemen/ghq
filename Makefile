VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X main.revision=$(CURRENT_REVISION)"
VERBOSE_FLAG = $(if $(VERBOSE),-v)
ifdef update
  u=-u
endif

export GO111MODULE=on

.PHONY: deps
deps:
	go get ${u} -d $(VERBOSE_FLAG)
	go mod tidy

.PHONY: devel-deps
devel-deps: deps
	GO111MODULE=off go get ${u} \
	  golang.org/x/lint/golint                  \
	  github.com/mattn/goveralls                \
	  github.com/Songmu/godzil/cmd/godzil       \
	  github.com/Songmu/goxz/cmd/goxz           \
	  github.com/Songmu/ghch/cmd/ghch           \
	  github.com/Songmu/gocredits/cmd/gocredits \
	  github.com/tcnksm/ghr

.PHONY: test
test: deps
	go test $(VERBOSE_FLAG) ./...

.PHONY: lint
lint: devel-deps
	go vet ./...
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

.PHONY: bump
bump: devel-deps
	godzil release

CREDITS: devel-deps go.sum
	gocredits -w

.PHONY: crossbuild
crossbuild: CREDITS
	rm -rf dist/snapshot
	cp ghq.txt README.txt
	goxz -arch=386,amd64 -build-ldflags=$(BUILD_LDFLAGS) \
      -include='zsh/_ghq' -z -d dist/snapshot
	cd dist/snapshot && shasum $$(find * -type f -maxdepth 0) > SHASUMS

.PHONY: upload
upload:
	ghr -body="$$(ghch --latest -F markdown)" v$(VERSION) dist/snapshot

.PHONY: release
release: bump docker-release

.PHONY: local-release
local-release: bump crossbuild upload

.PHONY: docker-release
docker-release:
	@docker run \
      -v $(PWD):/build \
      -w /build \
      -e GITHUB_TOKEN="$(GITHUB_TOKEN)" \
      --rm        \
      golang:1.12 \
      make crossbuild upload

ARCHIVE_DIR = ghq-$(VERSION)
.PHONY: archive
archive:
	@git archive HEAD --prefix=$(ARCHIVE_DIR)/ -o ghq-$(VERSION).tar
	@mkdir -p $(ARCHIVE_DIR)
	@echo $(CURRENT_REVISION) > $(ARCHIVE_DIR)/.revision
	@tar --append -vf ghq-$(VERSION).tar $(ARCHIVE_DIR)/.revision
	@gzip ghq-$(VERSION).tar
	@rm -rf $(ARCHIVE_DIR)
