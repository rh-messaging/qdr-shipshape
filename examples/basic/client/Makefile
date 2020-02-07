.PHONY: build
build:
	mvn clean package

.PHONY: image
image: build
	docker build -t qdrshipshape/examples-java-basic ./
