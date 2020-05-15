module github.com/weaveworks/wks

go 1.12

require (
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/alioygur/gores v1.2.1
	github.com/bitnami-labs/sealed-secrets v0.12.1 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/dlespiau/kube-test-harness v0.0.0-20190930170435-ec3f93e1a754 // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/fluxcd/go-git-providers v0.0.3
	github.com/ghodss/yaml v1.0.0
	github.com/gliderlabs/ssh v0.1.4 // indirect
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/google/go-cmp v0.4.0
	github.com/google/go-github/v32 v32.1.0
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/lithammer/dedent v1.1.0
	github.com/mattn/go-colorable v0.1.0 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/onsi/gomega v1.9.0 // indirect
	github.com/open-policy-agent/opa v0.12.2
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1 // indirect
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	github.com/sclevine/agouti v0.0.0-20190613051229-00c1187c74ad
	github.com/segmentio/ksuid v1.0.2
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	github.com/tebeka/selenium v0.9.9
	github.com/weaveworks/common v0.0.0-20190410110702-87611edc252e
	github.com/weaveworks/footloose v0.0.0-20190903132036-efbcbb7a6390
	github.com/weaveworks/wksctl v0.9.0-alpha.0
	github.com/xanzy/ssh-agent v0.2.1 // indirect
	github.com/yashtewari/glob-intersection v0.0.0-20180916065949-5c77d914dd0b // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/tools v0.0.0-20200804011535-6c149bb5ef0d
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a // indirect
	golang.org/x/sys v0.0.0-20200413165638-669c56c373c4 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/jcmturner/aescts.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/dnsutils.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/goidentity.v3 v3.0.0 // indirect
	gopkg.in/jcmturner/gokrb5.v6 v6.0.0-00010101000000-000000000000
	gopkg.in/jcmturner/rpc.v1 v1.1.0 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/src-d/go-git.v4 v4.10.0
	gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v3 v3.0.0-20200121175148-a6ecf24a6d71
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/yaml v1.2.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200403204345-e1beb1bd0f35 // indirect
	k8s.io/kubernetes v1.18.1 // indirect
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66 // indirect
	sigs.k8s.io/cluster-api v0.3.3 // indirect
	sigs.k8s.io/controller-runtime v0.5.2 // indirect
)

replace (
	github.com/appscode/jsonpatch => gomodules.xyz/jsonpatch/v2 v2.0.0
	gopkg.in/jcmturner/gokrb5.v6 => github.com/weaveworks/gokrb5 v0.0.0-20181126152309-94803fd23bf2
	github.com/docker/distribution => github.com/2opremio/distribution v0.0.0-20190419185413-6c9727e5e5de
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.1.0
	k8s.io/api => k8s.io/api v0.17.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.2
	k8s.io/apiserver => k8s.io/apiserver v0.17.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.2
	k8s.io/client-go => k8s.io/client-go v0.17.2
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.2
	k8s.io/code-generator => k8s.io/code-generator v0.17.2
	k8s.io/component-base => k8s.io/component-base v0.17.2
	k8s.io/cri-api => k8s.io/cri-api v0.17.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.2
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.2
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.2
	k8s.io/kubectl => k8s.io/kubectl v0.17.2
	k8s.io/kubelet => k8s.io/kubelet v0.17.2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.2
	k8s.io/metrics => k8s.io/metrics v0.17.2
	k8s.io/node-api => k8s.io/node-api v0.17.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.2
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.17.2
	k8s.io/sample-controller => k8s.io/sample-controller v0.17.2
)
