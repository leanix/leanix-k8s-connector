PROJECT ?= leanix-k8s-connector
DOCKER_NAMESPACE ?= leanixacrpublic.azurecr.io

VERSION := 6.5.1
VERSION_LATEST := 6.latest
FULL_VERSION := $(VERSION)-$(shell git describe --tags --always)

IMAGE := $(DOCKER_NAMESPACE)/$(PROJECT):$(VERSION)
IMAGE_LATEST := $(DOCKER_NAMESPACE)/$(PROJECT):$(VERSION_LATEST)
FULL_IMAGE := $(DOCKER_NAMESPACE)/$(PROJECT):$(FULL_VERSION)
LATEST := $(DOCKER_NAMESPACE)/$(PROJECT):latest
GOOS ?= linux
GOARCH ?= amd64

.PHONY: all

all: clean test build

clean:
	$(RM) bin/$(PROJECT)

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/$(PROJECT) -ldflags '-extldflags "-static"' ./cmd/$(PROJECT)/main.go

version:
	@echo $(VERSION)

gen:
	go mod download
	go install github.com/vektra/mockery/v2@v2.40.1
	mockery --all --recursive --with-expecter --case=underscore --output ./pkg/mocks

image:
	docker build --no-cache --pull --rm -t $(IMAGE) -t $(FULL_IMAGE) -t $(LATEST) -t $(IMAGE_LATEST) .

push:
	docker push $(IMAGE)
	docker push $(FULL_IMAGE)
	docker push $(LATEST)
	docker push $(IMAGE_LATEST)

test:
	go test ./pkg/...
