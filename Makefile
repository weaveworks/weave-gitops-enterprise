.PHONY: all install clean generated images lint unit-tests check
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
DOCKERFILES := $(shell find . -name tools -prune -o -name vendor -prune -o -name rpm -prune -o -name build -prune -o -name environments -prune -o -name test -prune -o -name examples -prune -o -type f -name 'Dockerfile' -print)
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

check: all lint unit-tests container-tests

BINARIES = \
	cmd/wksctl/wksctl \
	cmd/wks-entitle/wks-entitle \
	cmd/wks-ci/wks-ci \
	cmd/k8s-krb5-server/server \
	cmd/mock-authz-server/server \
	cmd/mock-https-authz-server/server \
	cmd/controller/controller \
	cmd/policy/policy \
	$(NULL)

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

MANIFESTS=$(shell find pkg/apis/wksprovider/controller/manifests/yaml -name '*.yaml' -print)
pkg/apis/wksprovider/controller/manifests/manifests_vfsdata.go: $(MANIFESTS)
	go generate ./pkg/apis/wksprovider/controller/manifests

CRDS=$(shell find pkg/apis/cluster-api/config/crds -name '*.yaml' -print)
pkg/apis/wksprovider/machine/os/crds_vfsdata.go: $(CRDS)
	go generate ./pkg/apis/wksprovider/machine/crds

POLICIES=$(shell find pkg/opa/policy/rego -name '*.rego' -print)
pkg/opa/policy/policy_vfsdata.go: $(POLICIES)
	go generate ./pkg/opa/policy

generated: pkg/guide/assets_vfsdata.go pkg/addons/assets/assets_vfsdata.go pkg/apis/wksprovider/controller/manifests/manifests_vfsdata.go pkg/apis/wksprovider/machine/scripts/scripts_vfsdata.go pkg/apis/wksprovider/machine/os/crds_vfsdata.go pkg/opa/policy/policy_vfsdata.go

cmd/wksctl/wksctl: $(DEPS) generated
cmd/wksctl/wksctl: cmd/wksctl/*.go
	CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION) -X main.imageTag=$(IMAGE_TAG)" -o $@ cmd/wksctl/*.go


cmd/policy/.uptodate: cmd/policy/policy
cmd/policy/policy: cmd/policy/*.go
	CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/policy/*.go

ENTITLE_DEPS=$(call godeps,./cmd/wks-entitle)
cmd/wks-entitle/wks-entitle: $(ENTITLE_DEPS)
	CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/wks-entitle/*.go

CI_DEPS=$(call godeps,./cmd/wks-ci)

cmd/wks-ci/.uptodate: cmd/wks-ci/wks-ci cmd/wks-ci/Dockerfile
cmd/wks-ci/wks-ci: $(CI_DEPS) cmd/wks-ci/*.go
	CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/wks-ci/*.go

cmd/controller/.uptodate: cmd/controller/controller cmd/controller/Dockerfile
cmd/controller/controller: $(DEPS) generated
cmd/controller/controller: cmd/controller/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/controller/*.go

cmd/k8s-krb5-server/.uptodate: cmd/k8s-krb5-server/server cmd/k8s-krb5-server/Dockerfile
cmd/k8s-krb5-server/server: cmd/k8s-krb5-server/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/k8s-krb5-server/*.go

cmd/mock-authz-server/.uptodate: cmd/mock-authz-server/server cmd/mock-authz-server/Dockerfile
cmd/mock-authz-server/server: cmd/mock-authz-server/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/mock-authz-server/*.go

cmd/mock-https-authz-server/.uptodate: cmd/mock-https-authz-server/server cmd/mock-https-authz-server/Dockerfile
cmd/mock-https-authz-server/server: cmd/mock-https-authz-server/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/mock-https-authz-server/*.go

install: all
	cp cmd/wksctl/wksctl `go env GOPATH`/bin
	cp cmd/wks-entitle/wks-entitle `go env GOPATH`/bin
	cp cmd/wks-ci/wks-ci `go env GOPATH`/bin

EMBEDMD_FILES = \
	docs/entitlements.md \
	$(NULL)

lint:
	@bin/go-lint
	@bin/check-embedmd.sh $(EMBEDMD_FILES)

clean:
	$(SUDO) docker rmi $(IMAGE_NAMES) >/dev/null 2>&1 || true
	rm -rf $(UPTODATE_FILES)
	go clean
	rm -f cmd/wksctl/wksctl
	rm -f cmd/controller/controller
	rm -f cmd/wks-ci/wks-ci

push:
	for IMAGE_NAME in $(IMAGE_NAMES); do \
		docker push $$IMAGE_NAME:$(IMAGE_TAG); \
	done

# We select which directory we want to descend into to not execute integration
# tests here.
unit-tests: generated
	go test -v ./cmd/... ./pkg/...

# Tests running in containers
mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(dir $(mkfile_path))

container-tests:  test/container/images/centos7/.uptodate pkg/apis/wksprovider/machine/scripts/scripts_vfsdata.go pkg/apis/wksprovider/controller/manifests/manifests_vfsdata.go
	go test -count=1 ./test/container/...

integration-tests-container: cmd/wksctl/wksctl cmd/controller/.uptodate
	IMAGE_TAG=$(IMAGE_TAG) go test -v -timeout 20m ./test/integration/container/...

FORCE:
