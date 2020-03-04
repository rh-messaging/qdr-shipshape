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

.PHONY: unit-tests
unit-tests:
	go test -v "./pkg/..."

.PHONY: images
images: build clients-python

.PHONY: clients-python
clients-python:
	docker build -t qdrshipshape/clients-python clients/python/
	docker tag qdrshipshape/clients-python docker.io/qdrshipshape/clients-python
	docker push docker.io/qdrshipshape/clients-python

.PHONY: travis-tests
travis-tests:
	ginkgo -v -r ./test/travis
