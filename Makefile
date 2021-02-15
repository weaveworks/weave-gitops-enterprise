.PHONY: all install clean images lint unit-tests check wksctl-version generate-manifests ui-build-for-tests
.DEFAULT_GOAL := all

# Boiler plate for bulding Docker containers.
# All this must go at top of file I'm afraid.
IMAGE_PREFIX := docker.io/weaveworks/wkp-
IMAGE_TAG := $(shell tools/image-tag)
GIT_REVISION := $(shell git rev-parse HEAD)
VERSION=$(shell git describe --always --match "v*")
CURRENT_DIR := $(shell pwd)
UPTODATE := .uptodate
BUILD_IN_CONTAINER=true
BUILD_IMAGE=docker.io/weaveworks/wkp-wks-build
BUILD_UPTODATE=wks-build/.uptodate
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

all: wksctl-version $(UPTODATE_FILES) binaries

check: all lint unit-tests container-tests

LOCAL_BINARIES = \
	cmd/wk/wk \
	cmd/wks-entitle/wks-entitle \
	cmd/update-manifest/update-manifest

BINARIES = \
	$(LOCAL_BINARIES) \
	cmd/git-provider-service/git-provider-service \
	cmd/gitops-repo-broker/gitops-repo-broker \
	cmd/mock-authz-server/server \
	cmd/mock-https-authz-server/server \
	cmd/ui-server/ui-server \
	cmd/wks-ci/checks/policy/policy \
	cmd/wkp-agent/wkp-agent \
	kerberos/cmd/k8s-krb5-server/server \
	kerberos/cmd/wk-kerberos/wk-kerberos \
	$(NULL)

binaries: $(BINARIES)

godeps=$(shell go list -deps -f '{{if not .Standard}}{{$$dep := .}}{{range .GoFiles}}{{$$dep.Dir}}/{{.}} {{end}}{{end}}' $1)

DEPS=$(call godeps,./cmd/wk)

GENERATED = pkg/guide/assets_vfsdata.go pkg/opa/policy/policy_vfsdata.go pkg/setup/setup_vfsdata.go

# .uptodate files are for Docker builds, which should happen outside of the container
cmd/wks-ci/checks/policy/.uptodate: cmd/policy/policy
cmd/wks-ci/.uptodate: cmd/wks-ci/wks-ci cmd/wks-ci/checks/policy/policy cmd/wks-ci/Dockerfile
kerberos/cmd/k8s-krb5-server/.uptodate: kerberos/cmd/k8s-krb5-server/server kerberos/cmd/k8s-krb5-server/Dockerfile
cmd/mock-authz-server/.uptodate: cmd/mock-authz-server/server cmd/mock-authz-server/Dockerfile
cmd/mock-https-authz-server/.uptodate: cmd/mock-https-authz-server/server cmd/mock-https-authz-server/Dockerfile
cmd/git-provider-service/.uptodate: cmd/git-provider-service/git-provider-service cmd/git-provider-service/Dockerfile
cmd/gitops-repo-broker/.uptodate: cmd/gitops-repo-broker/gitops-repo-broker cmd/gitops-repo-broker/Dockerfile
cmd/ui-server/.uptodate: cmd/ui-server/ui-server cmd/ui-server/Dockerfile cmd/ui-server/html
cmd/wkp-agent/.uptodate: cmd/wkp-agent/wkp-agent cmd/wkp-agent/Dockerfile

wkp-cluster-components/.uptodate: wkp-cluster-components/build

# Cluster Components
CC_CODE_DEPS = $(shell find wkp-cluster-components/src wkp-cluster-components/templates -type f)
CC_BUILD_DEPS = \
	wkp-cluster-components/.babelrc \
	wkp-cluster-components/package.json \
	wkp-cluster-components/package-lock.json
CC_DEPS = $(CC_CODE_DEPS) $(CC_BUILD_DEPS)
wkp-cluster-components/build: $(CC_DEPS)

# All dependencies for binaries must be listed outside of the BUILD_IN_CONTAINER if-statement.

USER_GUIDE_SOURCES=$(shell find user-guide/ -name public -prune -o -type f -print) user-guide/content/deps/_index.md
user-guide/public: $(USER_GUIDE_SOURCES)

# # Third-party build dependencies
# SCA_DEPS = \
# 	go.mod \
# 	ui/package.json \
# 	wkp-cluster-components/package.json \
# 	setup/wk-quickstart/setup/dependencies.toml \
# 	$(shell find wkp-cluster-components/templates -name 'helm-release.yaml' -prune -print)

# # Generate the third-party deps page for the user-guide if any of the deps have changed
# user-guide/content/deps/_index.md: $(SCA_DEPS)
# 	bin/sca-generate-deps.sh

pkg/guide/assets_vfsdata.go: user-guide/public

POLICIES=$(shell find pkg/opa/policy/rego -name '*.rego' -print)
pkg/opa/policy/policy_vfsdata.go: $(POLICIES)

SETUP=$(shell find setup -name bin -prune -o -type f -print)
pkg/setup/setup_vfsdata.go: $(SETUP)

cmd/wk/wk: $(DEPS) $(GENERATED)
cmd/wk/wk: cmd/wk/*.go

cmd/wks-ci/checks/policy/policy: cmd/wks-ci/checks/policy/*.go $(GENERATED)

ENTITLE_DEPS=$(call godeps,./cmd/wks-entitle)
cmd/wks-entitle/wks-entitle: $(ENTITLE_DEPS)

CI_DEPS=$(call godeps,./cmd/wks-ci)
cmd/wks-ci/wks-ci: $(CI_DEPS)

UPDATE_MANIFEST_DEPS=$(call godeps,./cmd/update-manifest)
cmd/update-manifest/update-manifest: $(UPDATE_MANIFEST_DEPS)

kerberos/cmd/wk-kerberos/wk-kerberos: $(call godeps,./kerberos/cmd/wk-kerberos/)
kerberos/cmd/k8s-krb5-server/server: kerberos/cmd/k8s-krb5-server/*.go
cmd/mock-authz-server/server: cmd/mock-authz-server/*.go
cmd/mock-https-authz-server/server: cmd/mock-https-authz-server/*.go
cmd/git-provider-service/git-provider-service: $(call godeps,./cmd/git-provider-service)
cmd/gitops-repo-broker/gitops-repo-broker: $(call godeps,./cmd/gitops-repo-broker)
cmd/ui-server/ui-server: cmd/ui-server/*.go

UI_CODE_DEPS = $(shell find ui/src -name '*.jsx' -or -name '*.json')
UI_BUILD_DEPS = \
	ui/.babelrc.js \
	ui/.eslintrc.js \
	ui/server.js \
	ui/yarn.lock \
	ui/webpack.common.js \
	ui/webpack.production.js
UI_DEPS = $(UI_CODE_DEPS) $(UI_BUILD_DEPS)
ui/build: $(UI_DEPS) user-guide/public

cmd/ui-server/html: ui/build
	cp -r ui/build $@

ifeq ($(BUILD_IN_CONTAINER),true)

$(BINARIES) $(GENERATED) wkp-cluster-components/build ui/build unit-tests generate-manifests lint: $(BUILD_UPTODATE)
	$(SUDO) docker run -ti --rm \
		-v $(shell pwd):/src/github.com/weaveworks/wks:delegated \
		-v $(shell go env GOPATH)/pkg:/go/pkg:delegated \
		--net=host \
		-e SRC_PATH=/src/github.com/weaveworks/wks -e GOPATH=/go/ \
		-e GOARCH -e GOOS -e CIRCLECI -e CIRCLE_BUILD_NUM -e CIRCLE_NODE_TOTAL \
		-e CIRCLE_NODE_INDEX -e COVERDIR -e SLOW -e TESTDIRS \
		$(BUILD_IMAGE) $@

else # not BUILD_IN_CONTAINER

user-guide/public:
	cd user-guide && ./make-static.sh $(VERSION)

pkg/guide/assets_vfsdata.go:
	go generate ./pkg/guide

pkg/opa/policy/policy_vfsdata.go:
	go generate ./pkg/opa/policy

pkg/setup/setup_vfsdata.go:
	RELEASE_GOOS=$(LOCAL_BINARIES_GOOS) ./tools/build/setup/build-release.sh $(CURRENT_DIR)/setup $(CURRENT_DIR)/setup/wk-quickstart/setup/dependencies.toml
	go generate ./pkg/setup
	# Clean up. FIXME: do this better.
	@rm -rf $(CURRENT_DIR)/setup/wk-quickstart/setup/VERSION
	@rm -rf $(CURRENT_DIR)/setup/wk-quickstart/.git

# ensure we use the same version for controller when go.mod references a wksctl release
WKSCTL_GO_MOD_VERSION=$(shell grep wksctl go.mod | cut -d' ' -f2 | egrep -v '=>|.*[-][0-9]{14}[-][0-9]{12}')
WKSCTL_DEPS_VERSION=$(shell awk 'BEGIN {FS="\""}; /\[controller]/ {found=1; next}; found==1 {print($$2); exit}' setup/wk-quickstart/setup/dependencies.toml)
wksctl-version:
	@test -n "$(WKSCTL_GO_MOD_VERSION)" && test "$(WKSCTL_GO_MOD_VERSION)" != "$(WKSCTL_DEPS_VERSION)" && ex '+/\[controller]' -c"+1|s/\".*\"/\"$(WKSCTL_GO_MOD_VERSION)\"/|x" setup/wk-quickstart/setup/dependencies.toml >/dev/null 2>&1 || true

cmd/wk/wk:
	CGO_ENABLED=0 GOOS=$(LOCAL_BINARIES_GOOS) GOARCH=amd64 go build -ldflags "-X github.com/weaveworks/wks/pkg/version.Version=$(VERSION) -X github.com/weaveworks/wks/pkg/version.ImageTag=$(IMAGE_TAG)" -o $@ cmd/wk/*.go

cmd/wks-ci/checks/policy/policy:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/wks-ci/checks/policy/*.go

cmd/wks-entitle/wks-entitle:
	CGO_ENABLED=0 GOOS=$(LOCAL_BINARIES_GOOS) GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/wks-entitle/*.go

cmd/wks-ci/wks-ci:
	CGO_ENABLED=0 GOOS=$(LOCAL_BINARIES_GOOS) GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/wks-ci/*.go

cmd/update-manifest/update-manifest:
	CGO_ENABLED=0 GOOS=$(LOCAL_BINARIES_GOOS) GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/update-manifest/*.go

kerberos/cmd/wk-kerberos/wk-kerberos:
	CGO_ENABLED=0 GOARCH=amd64 go build -o $@ ./kerberos/cmd/wk-kerberos

kerberos/cmd/k8s-krb5-server/server:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ ./kerberos/cmd/k8s-krb5-server

cmd/mock-authz-server/server:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/mock-authz-server/*.go

cmd/mock-https-authz-server/server:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/mock-https-authz-server/*.go

cmd/git-provider-service/git-provider-service:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ ./cmd/git-provider-service

cmd/gitops-repo-broker/gitops-repo-broker:
	CGO_ENABLED=1 GOARCH=amd64 go build -ldflags "-X github.com/weaveworks/wks/pkg/version.ImageTag=$(IMAGE_TAG) $(cgo_ldflags)" -o $@ ./cmd/gitops-repo-broker

cmd/wkp-agent/wkp-agent:
	CGO_ENABLED=0 GOOS=$(LOCAL_BINARIES_GOOS) GOARCH=amd64 go build -o $@ ./cmd/wkp-agent

# UI
cmd/ui-server/ui-server:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $@ cmd/ui-server/*.go

ui/build:
	cd ui && yarn install --frozen-lockfile && yarn lint && yarn test && yarn build
	cp -r user-guide/public ui/build/docs

wkp-cluster-components/build:
	cd wkp-cluster-components && \
		npm ci && \
		VERSION=$(VERSION) IMAGE_TAG=$(IMAGE_TAG) npm run build
	touch $@

generate-manifests:
	cd wkp-cluster-components && npm run generate-manifests

EMBEDMD_FILES = \
	docs/entitlements.md \
	$(NULL)

lint:
	bin/go-lint
	bin/check-embedmd.sh $(EMBEDMD_FILES)

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

endif # BUILD_IN_CONTAINER

ui-build-for-tests:
	cd ui && yarn install && yarn build

install: $(LOCAL_BINARIES)
	cp $(LOCAL_BINARIES) `go env GOPATH`/bin

clean:
	$(SUDO) docker rmi $(IMAGE_NAMES) >/dev/null 2>&1 || true
	$(SUDO) docker rmi $(patsubst %, %:$(IMAGE_TAG), $(IMAGE_NAMES)) >/dev/null 2>&1 || true
	rm -rf $(UPTODATE_FILES)
	rm -f $(BINARIES)
	rm -f $(GENERATED)
	rm -rf ui/build
	rm -rf cmd/ui-server/html

push:
	for IMAGE_NAME in $(IMAGE_NAMES); do \
		if ! echo $$IMAGE_NAME | grep build; then \
			docker push $$IMAGE_NAME:$(IMAGE_TAG); \
		fi \
	done

container-tests:
	WKP_DEBUG=true go test -count=1 ./test/container/...

cluster-component-tests: wkp-cluster-components/build
	WKP_DEBUG=true EXPECTED_VERSION=$(VERSION) EXPECTED_IMAGE_TAG=$(IMAGE_TAG) go test -v ./wkp-cluster-components/...


FORCE:
