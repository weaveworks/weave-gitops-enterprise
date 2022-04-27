.PHONY: all install clean images lint unit-tests check ui-build-for-tests update-mccp-chart-values ui-audit
.DEFAULT_GOAL := all

# Boiler plate for bulding Docker containers.
# All this must go at top of file I'm afraid.
IMAGE_PREFIX := docker.io/weaveworks/weave-gitops-enterprise-
IMAGE_TAG := $(shell tools/image-tag)
GIT_REVISION := $(shell git rev-parse HEAD)
VERSION=$(shell git describe --always --match "v*" --abbrev=7)
WEAVE_GITOPS_VERSION=$(shell git describe --always --match "v*" --abbrev=7 | sed 's/^[^0-9]*//')
TIME_NOW=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
CURRENT_DIR := $(shell pwd)
UPTODATE := .uptodate
GOOS := $(shell go env GOOS)
ifeq ($(GOOS),linux)
	cgo_ldflags=-linkmode external -w -extldflags "-static"
else
	# darwin doesn't like -static
	cgo_ldflags=-linkmode external -w
endif

# The GOOS to use for local binaries that we `make install`
LOCAL_BINARIES_GOOS ?= $(GOOS)

# Every directory with a Dockerfile in it builds an image called
# $(IMAGE_PREFIX)<dirname>. Dependencies (i.e. things that go in the image)
# still need to be explicitly declared.
%/$(UPTODATE): %/Dockerfile %/*
	$(SUDO) docker build \
		--build-arg=version=$(VERSION) \
		--build-arg=image_tag=$(IMAGE_TAG) \
		--build-arg=revision=$(GIT_REVISION) \
		--build-arg=now=$(TIME_NOW) \
		--tag $(IMAGE_PREFIX)$(shell basename $(@D)) \
		$(@D)/
	$(SUDO) docker tag $(IMAGE_PREFIX)$(shell basename $(@D)) $(IMAGE_PREFIX)$(shell basename $(@D)):$(IMAGE_TAG)
	touch $@

# Takes precedence over the more general rule above
# The only difference is the build context
cmd/event-writer/$(UPTODATE): cmd/event-writer/Dockerfile cmd/event-writer/*
	$(SUDO) docker build \
		--build-arg=version=$(VERSION) \
		--build-arg=image_tag=$(IMAGE_TAG) \
		--build-arg=revision=$(GIT_REVISION) \
		--build-arg=GITHUB_BUILD_TOKEN=$(GITHUB_BUILD_TOKEN) \
		--build-arg=now=$(TIME_NOW) \
		--tag $(IMAGE_PREFIX)$(shell basename $(@D)) \
		--file cmd/event-writer/Dockerfile \
        .
	$(SUDO) docker tag $(IMAGE_PREFIX)$(shell basename $(@D)) $(IMAGE_PREFIX)$(shell basename $(@D)):$(IMAGE_TAG)
	touch $@

# Takes precedence over the more general rule above
# The only difference is the build context
cmd/clusters-service/$(UPTODATE): cmd/clusters-service/Dockerfile cmd/clusters-service/clusters-service
	$(SUDO) docker build \
		--build-arg=version=$(WEAVE_GITOPS_VERSION) \
		--build-arg=image_tag=$(IMAGE_TAG) \
		--build-arg=revision=$(GIT_REVISION) \
		--build-arg=now=$(TIME_NOW) \
		--tag $(IMAGE_PREFIX)$(shell basename $(@D)) \
		--file cmd/clusters-service/Dockerfile \
		.
	$(SUDO) docker tag $(IMAGE_PREFIX)$(shell basename $(@D)) $(IMAGE_PREFIX)$(shell basename $(@D)):$(IMAGE_TAG)
	touch $@

WKP_AGENT := docker.io/weaveworks/wkp-agent
cmd/wkp-agent/$(UPTODATE): cmd/wkp-agent/Dockerfile cmd/wkp-agent/wkp-agent
	$(SUDO) docker build \
		--build-arg=version=$(WEAVE_GITOPS_VERSION) \
		--build-arg=image_tag=$(IMAGE_TAG) \
		--build-arg=revision=$(GIT_REVISION) \
		--build-arg=now=$(TIME_NOW) \
		--tag $(WKP_AGENT) \
		--file cmd/wkp-agent/Dockerfile \
		$(@D)/
	$(SUDO) docker tag $(WKP_AGENT) $(WKP_AGENT):$(IMAGE_TAG)
	touch $@

UI_SERVER := docker.io/weaveworks/weave-gitops-enterprise-ui-server
ui-cra/.uptodate: ui-cra/build
	$(SUDO) docker build \
		--build-arg=version=$(WEAVE_GITOPS_VERSION) \
		--build-arg=revision=$(GIT_REVISION) \
		--build-arg=GITHUB_TOKEN=$(GITHUB_BUILD_TOKEN) \
		--build-arg=now=$(TIME_NOW) \
		--tag $(UI_SERVER) \
		$(@D)/
	$(SUDO) docker tag $(UI_SERVER) $(UI_SERVER):$(IMAGE_TAG)
	touch $@

update-mccp-chart-values:
	sed -i "s|eventWriter: docker.io/weaveworks/weave-gitops-enterprise-event-writer.*|eventWriter: docker.io/weaveworks/weave-gitops-enterprise-event-writer:$(IMAGE_TAG)|" $(CHART_VALUES_PATH)
	sed -i "s|clustersService: docker.io/weaveworks/weave-gitops-enterprise-clusters-service.*|clustersService: docker.io/weaveworks/weave-gitops-enterprise-clusters-service:$(IMAGE_TAG)|" $(CHART_VALUES_PATH)
	sed -i "s|uiServer: docker.io/weaveworks/weave-gitops-enterprise-ui-server.*|uiServer: docker.io/weaveworks/weave-gitops-enterprise-ui-server:$(IMAGE_TAG)|" $(CHART_VALUES_PATH)

# Get a list of directories containing Dockerfiles
DOCKERFILES := $(shell find . \
	-name tools -prune -o \
	-name vendor -prune -o \
	-name rpm -prune -o \
	-name build -prune -o \
	-name environments -prune -o \
	-name test -prune -o \
	-name examples -prune -o \
	-name node_modules -prune -o \
	-name wks-ci -prune -o \
	-type f -name 'Dockerfile' -print)
UPTODATE_FILES := $(patsubst %/Dockerfile,%/$(UPTODATE),$(DOCKERFILES))
DOCKER_IMAGE_DIRS := $(patsubst %/Dockerfile,%,$(DOCKERFILES))
IMAGE_NAMES := $(foreach dir,$(DOCKER_IMAGE_DIRS),$(patsubst %,$(IMAGE_PREFIX)%,$(subst wkp-,,$(shell basename $(dir)))))
IMAGE_NAMES += $(WKP_AGENT)
images:
	$(info $(IMAGE_NAMES))
	@echo > /dev/null


# Define imagetag-golang, etc, for each image, which parses the dockerfile and
# prints an image tag. For example:
#     FROM golang:1.8.1-stretch
# in the "foo/Dockerfile" becomes:
#     $ make imagetag-foo
#     1.8.1-stretch
define imagetag_dep
.PHONY: imagetag-$(1)
$(patsubst $(IMAGE_PREFIX)%,imagetag-%,$(1)): $(patsubst $(IMAGE_PREFIX)%,%,$(1))/Dockerfile
	@cat $$< | grep "^FROM " | head -n1 | sed 's/FROM \(.*\):\(.*\)/\2/'
endef
$(foreach image, $(IMAGE_NAMES), $(eval $(call imagetag_dep, $(image))))

all: $(UPTODATE_FILES) binaries

check: all lint unit-tests ui-audit

BINARIES = \
	cmd/wkp-agent/wkp-agent \
	$(NULL)

binaries: $(BINARIES)

godeps=$(shell go list -deps -f '{{if not .Standard}}{{$$dep := .}}{{range .GoFiles}}{{$$dep.Dir}}/{{.}} {{end}}{{end}}' $1)

cmd/wkp-agent/wkp-agent:
	CGO_ENABLED=0 GOOS=$(LOCAL_BINARIES_GOOS) GOARCH=amd64 go build -o $@ ./cmd/wkp-agent

.PHONY: ui-cra/build
ui-cra/build:
	make build VERSION=$(VERSION) -C ui-cra

ui-audit:
	# Check js packages for any high or critical vulnerabilities
	cd ui-cra && yarn audit --level high; if [ $$? -gt 7 ]; then echo "Failed yarn audit"; exit 1; fi

lint:
	bin/go-lint

cmd/clusters-service/clusters-service: $(cmd find cmd/clusters-service -name '*.go') common/** pkg/**
	CGO_ENABLED=1 go build -ldflags "-X github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/version.Version=$(WEAVE_GITOPS_VERSION) -X github.com/weaveworks/weave-gitops-enterprise/pkg/version.ImageTag=$(IMAGE_TAG) $(cgo_ldflags)" -tags netgo -a -o $@ ./cmd/clusters-service

# We select which directory we want to descend into to not execute integration
# tests here.
unit-tests-with-coverage: $(GENERATED)
	go test -v -cover -coverprofile=.coverprofile ./cmd/... ./pkg/...
	cd cmd/event-writer && go test -v -cover -coverprofile=.coverprofile ./...
	cd common && go test -v -cover -coverprofile=.coverprofile ./...
	cd cmd/clusters-service && go test -v -cover -coverprofile=.coverprofile ./...

unit-tests: $(GENERATED)
	go test -v ./cmd/... ./pkg/...
	cd cmd/event-writer && go test -v ./...
	cd common && go test -v ./...
	cd cmd/clusters-service && go test -v ./...

ui-build-for-tests:
	# Github actions npm is slow sometimes, hence increasing the network-timeout 
	yarn config set network-timeout 300000 && cd ui-cra && yarn install && yarn build

clean:
	$(SUDO) docker rmi $(IMAGE_NAMES) >/dev/null 2>&1 || true
	$(SUDO) docker rmi $(patsubst %, %:$(IMAGE_TAG), $(IMAGE_NAMES)) >/dev/null 2>&1 || true
	rm -rf $(UPTODATE_FILES)
	rm -f $(BINARIES)
	rm -f $(GENERATED)
	rm -rf ui-cra/build

push:
	for IMAGE_NAME in $(IMAGE_NAMES); do \
		if ! echo $$IMAGE_NAME | grep build; then \
			docker push $$IMAGE_NAME:$(IMAGE_TAG); \
		fi \
	done

FORCE:
