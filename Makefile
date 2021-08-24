.PHONY: all install clean images lint unit-tests check ui-build-for-tests update-mccp-chart-values update-wkp-ui-chart-values
.DEFAULT_GOAL := all

# Boiler plate for bulding Docker containers.
# All this must go at top of file I'm afraid.
IMAGE_PREFIX := docker.io/weaveworks/wkp-
IMAGE_TAG := $(shell tools/image-tag)
GIT_REVISION := $(shell git rev-parse HEAD)
VERSION=$(shell git describe --always --match "v*")
WEAVE_GITOPS_VERSION=$(shell git describe --always --match "v*" | sed 's/^[^0-9]*//')
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
		--tag $(IMAGE_PREFIX)$(subst wkp-,,$(shell basename $(@D))) \
		$(@D)/
	$(SUDO) docker tag $(IMAGE_PREFIX)$(subst wkp-,,$(shell basename $(@D))) $(IMAGE_PREFIX)$(subst wkp-,,$(shell basename $(@D))):$(IMAGE_TAG)
	touch $@

# Takes precedence over the more general rule above
# The only difference is the build context
cmd/event-writer/$(UPTODATE): cmd/event-writer/Dockerfile cmd/event-writer/*
	$(SUDO) docker build \
		--build-arg=version=$(VERSION) \
		--build-arg=image_tag=$(IMAGE_TAG) \
		--build-arg=revision=$(GIT_REVISION) \
		--tag $(IMAGE_PREFIX)$(subst wkp-,,$(shell basename $(@D))) \
		--file cmd/event-writer/Dockerfile \
        .
	$(SUDO) docker tag $(IMAGE_PREFIX)$(subst wkp-,,$(shell basename $(@D))) $(IMAGE_PREFIX)$(subst wkp-,,$(shell basename $(@D))):$(IMAGE_TAG)
	touch $@

WEAVE_GITOPS_CLUSTERS_SERVICE := docker.io/weaveworks/weave-gitops-clusters-service
cmd/capi-server/$(UPTODATE): cmd/capi-server/Dockerfile cmd/capi-server/*
	$(SUDO) docker build \
		--build-arg=version=$(WEAVE_GITOPS_VERSION) \
		--build-arg=image_tag=$(IMAGE_TAG) \
		--build-arg=revision=$(GIT_REVISION) \
		--tag $(WEAVE_GITOPS_CLUSTERS_SERVICE) \
		--file cmd/capi-server/Dockerfile \
		.
	$(SUDO) docker tag $(WEAVE_GITOPS_CLUSTERS_SERVICE) $(WEAVE_GITOPS_CLUSTERS_SERVICE):$(IMAGE_TAG)
	touch $@

update-mccp-chart-values: update-wkp-ui-chart-values
	sed -i "s|gitopsRepoBroker: docker.io/weaveworks/wkp-gitops-repo-broker.*|gitopsRepoBroker: docker.io/weaveworks/wkp-gitops-repo-broker:$(IMAGE_TAG)|" $(CHART_VALUES_PATH)
	sed -i "s|eventWriter: docker.io/weaveworks/wkp-event-writer.*|eventWriter: docker.io/weaveworks/wkp-event-writer:$(IMAGE_TAG)|" $(CHART_VALUES_PATH)
	sed -i "s|clustersService: docker.io/weaveworks/weave-gitops-clusters-service.*|clustersService: docker.io/weaveworks/weave-gitops-clusters-service:$(IMAGE_TAG)|" $(CHART_VALUES_PATH)

update-wkp-ui-chart-values:
	sed -i "s|tag: .*|tag: $(IMAGE_TAG)|" $(CHART_VALUES_PATH)

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
IMAGE_NAMES += $(WEAVE_GITOPS_CLUSTERS_SERVICE)
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

check: all lint unit-tests

LOCAL_BINARIES = \
	cmd/mccp/mccp

BINARIES = \
	$(LOCAL_BINARIES) \
	cmd/gitops-repo-broker/gitops-repo-broker \
	cmd/ui-server/ui-server \
	cmd/wkp-agent/wkp-agent \
	$(NULL)

binaries: $(BINARIES)

godeps=$(shell go list -deps -f '{{if not .Standard}}{{$$dep := .}}{{range .GoFiles}}{{$$dep.Dir}}/{{.}} {{end}}{{end}}' $1)

# .uptodate files are for Docker builds, which should happen outside of the container
cmd/gitops-repo-broker/.uptodate: cmd/gitops-repo-broker/gitops-repo-broker cmd/gitops-repo-broker/Dockerfile
cmd/ui-server/.uptodate: cmd/ui-server/ui-server cmd/ui-server/Dockerfile cmd/ui-server/html
cmd/wkp-agent/.uptodate: cmd/wkp-agent/wkp-agent cmd/wkp-agent/Dockerfile

cmd/mccp/mccp: cmd/mccp/*.go

cmd/gitops-repo-broker/gitops-repo-broker: $(call godeps,./cmd/gitops-repo-broker)
cmd/ui-server/ui-server: cmd/ui-server/*.go

cmd/ui-server/html: ui-cra/build
	mkdir -p $@
	cp -r ui-cra/build $@/mccp

cmd/mccp/mccp:
	CGO_ENABLED=0 GOOS=$(LOCAL_BINARIES_GOOS) GOARCH=amd64 go build -ldflags "-X github.com/weaveworks/weave-gitops-enterprise/pkg/version.Version=$(VERSION) -X github.com/weaveworks/weave-gitops-enterprise/pkg/version.ImageTag=$(IMAGE_TAG)" -o $@ cmd/mccp/*.go

cmd/gitops-repo-broker/gitops-repo-broker:
	CGO_ENABLED=1 GOARCH=amd64 go build -ldflags "-X github.com/weaveworks/weave-gitops-enterprise/pkg/version.ImageTag=$(IMAGE_TAG) $(cgo_ldflags)" -o $@ ./cmd/gitops-repo-broker

cmd/wkp-agent/wkp-agent:
	CGO_ENABLED=0 GOOS=$(LOCAL_BINARIES_GOOS) GOARCH=amd64 go build -o $@ ./cmd/wkp-agent

# UI
cmd/ui-server/ui-server:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/ui-server/*.go

ui-cra/build:
	cd ui-cra && yarn install --frozen-lockfile && yarn build

lint:
	bin/go-lint

# We select which directory we want to descend into to not execute integration
# tests here.
unit-tests-with-coverage: $(GENERATED)
	WKP_DEBUG=true go test -cover -coverprofile=.coverprofile ./cmd/... ./pkg/...
	cd cmd/event-writer && go test -cover -coverprofile=.coverprofile ./converter/... ./database/... ./liveness/... ./subscribe/... ./run/... ./test/...
	cd common && go test -cover -coverprofile=.coverprofile ./...

unit-tests: $(GENERATED)
	WKP_DEBUG=true go test -v ./cmd/... ./pkg/...
	cd cmd/event-writer && go test ./converter/... ./database/... ./liveness/... ./subscribe/... ./run/... ./test/...
	cd common && go test ./...

ui-build-for-tests:
	cd ui-cra && yarn install && yarn build

install: $(LOCAL_BINARIES)
	cp $(LOCAL_BINARIES) `go env GOPATH`/bin

clean:
	$(SUDO) docker rmi $(IMAGE_NAMES) >/dev/null 2>&1 || true
	$(SUDO) docker rmi $(patsubst %, %:$(IMAGE_TAG), $(IMAGE_NAMES)) >/dev/null 2>&1 || true
	rm -rf $(UPTODATE_FILES)
	rm -f $(BINARIES)
	rm -f $(GENERATED)
	rm -rf ui-cra/build
	rm -rf cmd/ui-server/html

push:
	for IMAGE_NAME in $(IMAGE_NAMES); do \
		if ! echo $$IMAGE_NAME | grep build; then \
			docker push $$IMAGE_NAME:$(IMAGE_TAG); \
		fi \
	done

FORCE:
