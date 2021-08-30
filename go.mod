module github.com/weaveworks/weave-gitops-enterprise

go 1.16

require (
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.3.1
	github.com/cloudevents/sdk-go/v2 v2.3.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/go-logfmt/logfmt v0.5.0
	github.com/go-openapi/strfmt v0.19.5 // indirect
	github.com/go-openapi/validate v0.19.8 // indirect
	github.com/go-playground/validator/v10 v10.4.1
	github.com/go-resty/resty/v2 v2.5.0
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0
	github.com/jarcoal/httpmock v1.0.8
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/nats-io/nats-server/v2 v2.1.7
	github.com/nats-io/nats.go v1.10.0
	github.com/nats-io/nkeys v0.1.4 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/alertmanager v0.21.0
	github.com/sclevine/agouti v0.0.0-20190613051229-00c1187c74ad
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tebeka/selenium v0.9.9
	github.com/weaveworks/cluster-api-provider-existinginfra v0.2.5
	github.com/weaveworks/common v0.0.0-20190410110702-87611edc252e
	github.com/weaveworks/footloose v0.0.0-20210208164054-2862489574a3
	github.com/weaveworks/libgitops v0.0.2
	github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server v0.0.0-00010101000000-000000000000
	github.com/weaveworks/weave-gitops-enterprise/common v0.0.0
	github.com/weaveworks/wksctl v0.10.2
	github.com/xanzy/go-gitlab v0.43.0 // indirect
	golang.org/x/tools v0.1.3
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gorm.io/datatypes v1.0.0
	gorm.io/gorm v1.21.11
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/cli-runtime v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/kubernetes v1.21.1 // indirect
	sigs.k8s.io/cluster-api v0.3.16
	sigs.k8s.io/controller-runtime v0.9.1
	sigs.k8s.io/kustomize/kyaml v0.11.0
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.7.0
	github.com/appscode/jsonpatch => gomodules.xyz/jsonpatch/v2 v2.0.0
	github.com/docker/distribution => github.com/2opremio/distribution v0.0.0-20190419185413-6c9727e5e5de
	github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server => ./cmd/capi-server
	github.com/weaveworks/weave-gitops-enterprise/common => ./common
	gopkg.in/jcmturner/gokrb5.v6 => github.com/weaveworks/gokrb5 v0.0.0-20181126152309-94803fd23bf2
	k8s.io/api => k8s.io/api v0.21.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.1
	k8s.io/apiserver => k8s.io/apiserver v0.21.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.1
	k8s.io/client-go => k8s.io/client-go v0.21.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.1
	k8s.io/code-generator => k8s.io/code-generator v0.21.1
	k8s.io/component-base => k8s.io/component-base v0.21.1
	k8s.io/component-helpers => k8s.io/component-helpers v0.21.1
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.1
	k8s.io/cri-api => k8s.io/cri-api v0.21.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.1
	k8s.io/kubectl => k8s.io/kubectl v0.21.1
	k8s.io/kubelet => k8s.io/kubelet v0.21.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.1
	k8s.io/metrics => k8s.io/metrics v0.21.1
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.3-rc.0
	k8s.io/node-api => k8s.io/node-api v0.21.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.1
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.21.1
	k8s.io/sample-controller => k8s.io/sample-controller v0.21.1
)
