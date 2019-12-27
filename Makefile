.PHONY: all
all: build

.PHONY: format
format:
	go fmt ./pkg/...

.PHONY: build
build:
	go build ./...

.PHONY: all-tests
all-tests: build
	ginkgo -v -r ./test

.PHONY: smoke-tests
smoke-tests: build
	ginkgo -v -r ./test/smoke

.PHONY: travis-tests
travis-tests:
	ginkgo -v -r ./test/travis
