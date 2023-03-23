GOOS := $(shell go env GOOS)
ifeq ($(GOOS),linux)
    cgo_ldflags=-linkmode external -w -extldflags "-static"
else
    # darwin doesn't like -static
    cgo_ldflags=-linkmode external -w
endif


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
	CGO_ENABLED=1 go build -ldflags "$(cgo_ldflags)" -a -installsuffix cgo -o bin/collector main.go

.PHONY: build-linux
build-linux:
# set CC and CXX cross compile vars for darwin
ifeq ($(shell uname),Darwin)
	CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ GOARCH=amd64 GOOS=linux CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" -a -installsuffix cgo -o bin/collector main.go
else
	GOARCH=amd64 GOOS=linux CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" -a -installsuffix cgo -o bin/collector main.go
endif