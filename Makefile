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

DIST_DIR = dist/v$(VERSION)
.PHONY: crossbuild
crossbuild: CREDITS
	rm -rf $(DIST_DIR)
	goxz -arch=386,amd64 -build-ldflags=$(BUILD_LDFLAGS) \
      -include='zsh/_ghq' -z -d $(DIST_DIR)
	cd $(DIST_DIR) && shasum $$(find * -type f -maxdepth 0) > SHASUMS

.PHONY: upload
upload:
	ghr -body="$$(ghch --latest -F markdown)" v$(VERSION) $(DIST_DIR)

.PHONY: release
release: bump docker-release

.PHONY: local-release
local-release: bump crossbuild upload

.PHONY: docker-release
docker-release:
	@docker run \
      -v $(PWD):/ghq \
      -w /ghq \
      -e GITHUB_TOKEN="$(GITHUB_TOKEN)" \
      --rm        \
      golang:1.13 \
      make crossbuild upload
