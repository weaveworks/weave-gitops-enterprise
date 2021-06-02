module github.com/weaveworks/wks/cmd/capi-server

go 1.16

require (
	github.com/google/go-cmp v0.5.6
	github.com/gorilla/mux v1.8.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v1.0.0
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	sigs.k8s.io/cluster-api v0.3.16
	sigs.k8s.io/yaml v1.2.0
)
