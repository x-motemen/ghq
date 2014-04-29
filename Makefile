VERBOSE_FLAG = $(if $(VERBOSE),-v)

build: deps
	go build $(VERBOSE_FLAG) ./...

test:
	go test $(VERBOSE_FLAG) ./...

deps:
	go get $(VERBOSE_FLAG) ./...

.PHONY: build test deps
