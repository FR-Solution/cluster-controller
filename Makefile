release = $(shell cat .release)
tag = $(DOCKER_USER)/cluster-controller:$(release)

build-and-push:
	docker build -t $(tag) --build-arg VERSION=$(release) -f Dockerfile .
	docker image push $(tag)