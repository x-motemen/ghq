VERBOSE_FLAG = $(if $(VERBOSE),-v)

BUILD_FLAGS = -ldflags "-X main.VERSION $$(git describe --tags --always --dirty)"

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

.PHONY: build test deps testdeps install
