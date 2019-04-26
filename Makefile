GO = go

VERBOSE_FLAG = $(if $(VERBOSE),-v)

VERSION = $$(git describe --tags --always --dirty) ($$(git name-rev --name-only HEAD))

BUILD_FLAGS = -ldflags "\
	      -X \"main.Version=$(VERSION)\" \
	      "

build: deps
	$(GO) build $(VERBOSE_FLAG) $(BUILD_FLAGS)

test: testdeps
	$(GO) test $(VERBOSE_FLAG) $($(GO) list ./... | grep -v '^github.com/motemen/ghq/vendor/')

deps:
	$(GO) get -d $(VERBOSE_FLAG)

testdeps:
	$(GO) get -d -t $(VERBOSE_FLAG)

install: deps
	$(GO) install $(VERBOSE_FLAG) $(BUILD_FLAGS)

bump-minor:
	git diff --quiet && git diff --cached --quiet
	new_version=$$(gobump minor -w -r -v) && \
	test -n "$$new_version" && \
	git commit -a -m "bump version to $$new_version" && \
	git tag v$$new_version

.PHONY: build test deps testdeps install
