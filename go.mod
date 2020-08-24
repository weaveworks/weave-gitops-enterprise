module github.com/weaveworks/wks

go 1.12

require (
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/alioygur/gores v1.2.1
	github.com/blang/semver v3.5.1+incompatible
	github.com/docker/distribution v2.7.1+incompatible
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gliderlabs/ssh v0.1.4 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/google/go-cmp v0.4.0
	github.com/google/go-github/v26 v26.1.3
	github.com/google/go-github/v28 v28.1.1
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lithammer/dedent v1.1.0
	github.com/mattn/go-colorable v0.1.0 // indirect
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/open-policy-agent/opa v0.12.2
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.8.1
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	github.com/sclevine/agouti v0.0.0-20190613051229-00c1187c74ad
	github.com/segmentio/ksuid v1.0.2
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.5.1
	github.com/tebeka/selenium v0.9.9
	github.com/weaveworks/common v0.0.0-20190410110702-87611edc252e
	github.com/weaveworks/footloose v0.0.0-20190903132036-efbcbb7a6390
	github.com/weaveworks/wksctl v0.8.5-rc.1
	github.com/xanzy/ssh-agent v0.2.1 // indirect
	github.com/yashtewari/glob-intersection v0.0.0-20180916065949-5c77d914dd0b // indirect
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	golang.org/x/tools v0.0.0-20190909030654-5b82db07426d
	gopkg.in/jcmturner/aescts.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/dnsutils.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/goidentity.v3 v3.0.0 // indirect
	gopkg.in/jcmturner/gokrb5.v6 v6.0.0-00010101000000-000000000000
	gopkg.in/jcmturner/rpc.v1 v1.1.0 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/src-d/go-git.v4 v4.10.0
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20191026110619-0b21df46bc1d
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.0.0-20190831074750-7364b6bdad65
	k8s.io/apimachinery v0.0.0-20190831074630-461753078381
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v1.0.0 // indirect
	sigs.k8s.io/cluster-api v0.0.0-20181211193542-3547f8dd9307
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	github.com/appscode/jsonpatch => gomodules.xyz/jsonpatch/v2 v2.0.0+incompatible
	github.com/dlespiau/kube-test-harness => github.com/dlespiau/kube-test-harness v0.0.0-20180712150055-7eab798dff48
	github.com/json-iterator/go => github.com/json-iterator/go v0.0.0-20180612202835-f2b4162afba3
	gopkg.in/jcmturner/gokrb5.v6 => github.com/weaveworks/gokrb5 v0.0.0-20181126152309-94803fd23bf2
	k8s.io/api => k8s.io/api v0.0.0-20190704094930-781da4e7b28a
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190704094625-facf06a8f4b8
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190202011228-6e4752048fde
	k8s.io/kubernetes => k8s.io/kubernetes v1.13.9-beta.0.0.20190726214758-e065364bfbf4
	sigs.k8s.io/kind => sigs.k8s.io/kind v0.0.0-20190204012257-d1773a79317d
)
