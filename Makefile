VERBOSE_FLAG = $(if $(VERBOSE),-v)

VERSION = $$(git describe --tags --always --dirty) ($$(git name-rev --name-only HEAD))

BUILD_FLAGS = -ldflags "\
	      -X \"main.Version=$(VERSION)\" \
	      "

build: deps
	go build $(VERBOSE_FLAG) $(BUILD_FLAGS)

test: testdeps
	go test $(VERBOSE_FLAG) ./...

deps:
	go get -d $(VERBOSE_FLAG)

testdeps:
	go get -d -t $(VERBOSE_FLAG)

install: deps
	go install $(VERBOSE_FLAG) $(BUILD_FLAGS)

bump-minor:
	git diff --quiet && git diff --cached --quiet
	new_version=$$(gobump minor -w -r -v) && \
	test -n "$$new_version" && \
	git commit -a -m "bump version to $$new_version" && \
	git tag v$$new_version

.PHONY: build test deps testdeps install
