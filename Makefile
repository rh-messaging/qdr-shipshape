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
	docker build -t qdrshipshape/clients-python:dispatch-1626 clients/python/
	docker tag qdrshipshape/clients-python:dispatch-1626 docker.io/qdrshipshape/clients-python:dispatch-1626
	docker push docker.io/qdrshipshape/clients-python:dispatch-1626

.PHONY: travis-tests
travis-tests:
	ginkgo -v -r ./test/travis
