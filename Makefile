VERBOSE_FLAG = $(if $(VERBOSE),-v)

build: deps
	go build $(VERBOSE_FLAG)

test: testdeps
	go test $(VERBOSE_FLAG) ./...

deps:
	go get $(VERBOSE_FLAG) ./...

testdeps:
	go get -t $(VERBOSE_FLAG) ./...

.PHONY: build test deps testdeps
