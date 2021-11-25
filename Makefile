PROJECT ?= leanix-k8s-connector
DOCKER_NAMESPACE ?= leanixacrpublic.azurecr.io

VERSION := 6.0.2
FULL_VERSION := $(VERSION)-$(shell git describe --tags --always)

IMAGE := $(DOCKER_NAMESPACE)/$(PROJECT):$(VERSION)
FULL_IMAGE := $(DOCKER_NAMESPACE)/$(PROJECT):$(FULL_VERSION)
LATEST := $(DOCKER_NAMESPACE)/$(PROJECT):latest
GOOS ?= linux
GOARCH ?= amd64

.PHONY: all

all: clean test build

clean:
	$(RM) bin/$(PROJECT)

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/$(PROJECT) -ldflags '-X $(shell go list -m)/pkg/version.VERSION=${FULL_VERSION} -extldflags "-static"' ./cmd/$(PROJECT)/main.go

version:
	@echo $(VERSION)

image:
	docker build -t $(IMAGE) -t $(FULL_IMAGE) -t $(LATEST) .

push:
	docker push $(IMAGE)
	docker push $(FULL_IMAGE)
	docker push $(LATEST)

test:
	go test ./pkg/...
