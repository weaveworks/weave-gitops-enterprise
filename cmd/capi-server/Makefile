.PHONY: generate
generate:
	@buf generate
	oapi-codegen -config oapi-codegen.config.yaml api/capi_server.swagger.json

.PHONY: install
install:
	@go install \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts \
		github.com/deepmap/oapi-codegen/cmd/oapi-codegen

.PHONY: build
build:
	go build -o capi-server main.go

