module github.com/weaveworks/wks/cmd/capi-server

go 1.16

require (
	github.com/fluxcd/go-git-providers v0.2.0
	github.com/google/go-cmp v0.5.6
	github.com/google/go-github/v32 v32.1.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.4.0
	github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts v1.1.1
	github.com/mkmik/multierror v0.3.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/weaveworks/wks/common v0.0.0-00010101000000-000000000000
	github.com/xanzy/go-gitlab v0.43.0
	golang.org/x/oauth2 v0.0.0-20210427180440-81ed05c6b58c
	golang.org/x/sys v0.0.0-20210608053332-aa57babbf139 // indirect
	google.golang.org/genproto v0.0.0-20210601170153-0befbe3492e2
	google.golang.org/grpc v1.38.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
	gorm.io/gorm v1.21.11
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/klog/v2 v2.10.0 // indirect
	sigs.k8s.io/cluster-api v0.3.16
	sigs.k8s.io/controller-runtime v0.9.0
	sigs.k8s.io/kustomize/kyaml v0.11.0
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/weaveworks/wks/common => ../../common
