.PHONY: all install clean gen images lint unit-tests
.DEFAULT_GOAL := all

# Boiler plate for bulding Docker containers.
# All this must go at top of file I'm afraid.
IMAGE_PREFIX := quay.io/wks/
IMAGE_TAG := $(shell tools/image-tag)
GIT_REVISION := $(shell git rev-parse HEAD)
VERSION=$(shell git describe)
UPTODATE := .uptodate

# Every directory with a Dockerfile in it builds an image called
# $(IMAGE_PREFIX)<dirname>. Dependencies (i.e. things that go in the image)
# still need to be explicitly declared.
%/$(UPTODATE): %/Dockerfile %/*
	$(SUDO) docker build --build-arg=revision=$(GIT_REVISION) -t $(IMAGE_PREFIX)$(shell basename $(@D)) $(@D)/
	$(SUDO) docker tag $(IMAGE_PREFIX)$(shell basename $(@D)) $(IMAGE_PREFIX)$(shell basename $(@D)):$(IMAGE_TAG)
	touch $@

# Get a list of directories containing Dockerfiles
DOCKERFILES := $(shell find . -name tools -prune -o -name vendor -prune -o -name rpm -prune -o -name build -prune -o -name environments -prune -o -name test -prune -o -type f -name 'Dockerfile' -print)
UPTODATE_FILES := $(patsubst %/Dockerfile,%/$(UPTODATE),$(DOCKERFILES))
DOCKER_IMAGE_DIRS := $(patsubst %/Dockerfile,%,$(DOCKERFILES))
IMAGE_NAMES := $(foreach dir,$(DOCKER_IMAGE_DIRS),$(patsubst %,$(IMAGE_PREFIX)%,$(shell basename $(dir))))
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

binaries: cmd/wksctl/wksctl cmd/k8s-krb5-server/server cmd/mock-authz-server/server cmd/controller/controller

godeps=$(shell go list -f '{{join .Deps "\n"}}' $1 | \
	   grep -v /vendor/ | \
	   xargs go list -f \
	   '{{if not .Standard}}{{ $$dep := . }}{{range .GoFiles}}{{$$dep.Dir}}/{{.}} {{end}}{{end}}')

DEPS=$(call godeps,./cmd/wksctl)

USER_GUIDE_SOURCES=$(shell find user-guide/ -name public -prune -o -print)
user-guide/public: $(USER_GUIDE_SOURCES)
	cd user-guide && ./make-static.sh

pkg/guide/assets_vfsdata.go: user-guide/public
	go generate ./pkg/guide

ADDONS_SOURCES=$(shell find addons/ -print)
pkg/addons/assets/assets_vfsdata.go: $(ADDONS_SOURCES)
	go generate ./pkg/addons/assets

SCRIPTS=$(shell find pkg/apis/wksprovider/machine/scripts/all -name '*.sh' -print)
pkg/apis/wksprovider/machine/scripts/scripts_vfsdata.go: $(SCRIPTS)
	go generate ./pkg/apis/wksprovider/machine/scripts

ALL_ASSETS = pkg/guide/assets_vfsdata.go pkg/addons/assets/assets_vfsdata.go pkg/apis/wksprovider/machine/scripts/scripts_vfsdata.go

cmd/wksctl/wksctl: $(DEPS) $(ALL_ASSETS)
cmd/wksctl/wksctl: cmd/wksctl/*.go
	CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/wksctl/*.go

cmd/controller/.uptodate: cmd/controller/controller cmd/controller/Dockerfile
cmd/controller/controller: $(DEPS) $(ALL_ASSETS)
cmd/controller/controller: cmd/controller/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/controller/*.go

cmd/k8s-krb5-server/.uptodate: cmd/k8s-krb5-server/server cmd/k8s-krb5-server/Dockerfile
cmd/k8s-krb5-server/server: cmd/k8s-krb5-server/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/k8s-krb5-server/*.go

cmd/mock-authz-server/.uptodate: cmd/mock-authz-server/server cmd/mock-authz-server/Dockerfile
cmd/mock-authz-server/server: cmd/mock-authz-server/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/mock-authz-server/*.go

install: cmd/wksctl/wksctl
	cp $^ `go env GOPATH`/bin

EMBEDMD_FILES = \
	docs/entitlements.md \
	$(NULL)

lint:
	@bin/go-lint
	@bin/check-embedmd.sh $(EMBEDMD_FILES)

gen:
	go install ./vendor/k8s.io/code-generator/cmd/deepcopy-gen
	deepcopy-gen \
		-i ./pkg/baremetalproviderspec/v1alpha1,./pkg/baremetalproviderspec \
		-O zz_generated.deepcopy \
		-h boilerplate.go.txt
	bin/embedmd.sh $(EMBEDMD_FILES)

clean:
	$(SUDO) docker rmi $(IMAGE_NAMES) >/dev/null 2>&1 || true
	rm -rf $(UPTODATE_FILES)
	go clean
	rm -f cmd/wksctl/wksctl
	rm -f cmd/controller/controller

push:
	for IMAGE_NAME in $(IMAGE_NAMES); do \
		docker push $$IMAGE_NAME:$(IMAGE_TAG); \
	done

# We select which directory we want to descend into to not execute integration
# tests here.
unit-tests:
	go test -v ./cmd/... ./pkg/...

# Tests running in containers
test/resource/tests: FORCE
	go test -c -o $@ ./test/resource

container-tests:  test/resource/tests test/images/centos7/.uptodate
	./test/run-in-docker.sh $< -test.v

# Integration tests, requiring to provision VMs
integration-test:
	go test -failfast -v -timeout 1h ./test -args -run.interactive -cmd /tmp/workspace/cmd/wksctl/wksctl \
			-tags.wks-k8s-krb5-server=$(IMAGE_TAG) \
			-tags.wks-mock-authz-server=$(IMAGE_TAG)

FORCE:
