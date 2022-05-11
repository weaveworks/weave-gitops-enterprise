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
	go build -o clusters-service main.go

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

crd-manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./api/capi/... ./api/tfcontroller/..." output:crd:artifacts:config=config/crd/bases
	cp config/crd/bases/* ../../charts/mccp/crds/

crd-generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./api/capi/... ./api/tfcontroller/..."

# To regenerate the CRD after changes to api/v1/capitemplate_types.go, run make crd-generate
# Note that controller-gen needs to be in $PATH
CONTROLLER_GEN = controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

.PHONY: unit-tests
unit-tests:
	@go test -v ./...

# Runs all tests, including integration tests
.PHONY: tests
tests:
	@go test -v ./... -tags=integration