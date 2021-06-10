module github.com/weaveworks/wks/cmd/capi-server

go 1.16

require (
	github.com/google/go-cmp v0.5.6
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.4.0
	github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts v1.1.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.0.1-0.20201006035406-b97b5ead31f7
	google.golang.org/genproto v0.0.0-20210601170153-0befbe3492e2
	google.golang.org/grpc v1.38.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.26.0
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	sigs.k8s.io/cluster-api v0.3.16
	sigs.k8s.io/controller-runtime v0.5.14
	sigs.k8s.io/yaml v1.2.0
)
