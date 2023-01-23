CONTROLLER_TOOLS_VERSION ?= v0.9.2
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
GOOS := $(shell go env GOOS)


## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: generate
generate:
	@buf generate

.PHONY: install
install:
	@go install \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts

.PHONY: build
build:
	go build -a -o bin/clusters-service main.go

.PHONY: build-linux
build-linux:
# set CC and CXX cross compile vars for darwin
ifeq ($(shell uname),Darwin)
	CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ GOARCH=amd64 GOOS=linux go build -ldflags "-linkmode external -extldflags -static" -a -o bin/clusters-service main.go
else
	GOARCH=amd64 GOOS=linux go build -ldflags "-linkmode external -extldflags -static" -a -o bin/clusters-service main.go
endif

.PHONY: crd-manifests
crd-manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./api/..." output:crd:artifacts:config=config/crd/bases
	cp  config/crd/bases/* ../../charts/mccp/crds/

.PHONY: crd-generate
crd-generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./api/..."

.PHONY: unit-tests
unit-tests:
	@go test -v ./...

# Runs all tests, including integration tests
.PHONY: tests
tests:
	@go test -v ./... -tags=integration

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
