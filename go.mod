module github.com/weaveworks/wks

go 1.12

require (
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitnami-labs/sealed-secrets v0.7.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/cavaliercoder/go-rpm v0.0.0-20190131055624-7a9c54e3d83e
	github.com/dlespiau/kube-test-harness v0.0.0-20180712150055-7eab798dff48
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/fatih/structs v1.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/gliderlabs/ssh v0.1.4 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/google/go-cmp v0.3.0
	github.com/google/go-github/v26 v26.1.3
	github.com/google/go-jsonnet v0.11.2
	github.com/hashicorp/go-uuid v0.0.0-20180228145832-27454136f036 // indirect
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/jcmturner/gofork v0.0.0-20180107083740-2aebee971930 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lithammer/dedent v1.1.0
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/onsi/gomega v1.4.2 // indirect
	github.com/open-policy-agent/opa v0.12.2
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.4 // indirect
	github.com/prometheus/procfs v0.0.3 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	github.com/segmentio/ksuid v1.0.2
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	github.com/thanhpk/randstr v0.0.0-20190104161604-ac5b2d62bffb
	github.com/weaveworks/common v0.0.0-20180717091316-5d6891c6875f
	github.com/weaveworks/footloose v0.0.0-20190829132911-efbcbb7a6390
	github.com/weaveworks/wksctl v0.0.0-20191001112524-9d38d9e2d69d
	github.com/xanzy/ssh-agent v0.2.1 // indirect
	github.com/yashtewari/glob-intersection v0.0.0-20180916065949-5c77d914dd0b // indirect
	golang.org/x/crypto v0.0.0-20190617133340-57b3e21c3d56
	golang.org/x/tools v0.0.0-20190816200558-6889da9d5479
	gopkg.in/jcmturner/aescts.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/dnsutils.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/goidentity.v3 v3.0.0 // indirect
	gopkg.in/jcmturner/gokrb5.v6 v6.0.5
	gopkg.in/jcmturner/rpc.v1 v1.1.0 // indirect
	gopkg.in/oleiade/reflections.v1 v1.0.0
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/src-d/go-git.v4 v4.10.0
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190831074750-7364b6bdad65
	k8s.io/apimachinery v0.0.0-20190831074630-461753078381
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cluster-bootstrap v0.0.0-20190205054431-5627c5c14d7e
	k8s.io/kube-proxy v0.0.0-20190208174132-30e63035f31f
	k8s.io/kubernetes v0.0.0-20190201210629-c6d339953bd4
	k8s.io/utils v0.0.0-20190801114015-581e00157fb1
	sigs.k8s.io/cluster-api v0.0.0-20181211193542-3547f8dd9307
	sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/yaml v1.1.0
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
