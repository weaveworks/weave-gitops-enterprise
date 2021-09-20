module github.com/weaveworks/weave-gitops-enterprise

go 1.17

require (
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.3.1
	github.com/cloudevents/sdk-go/v2 v2.3.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/go-logfmt/logfmt v0.5.0
	github.com/go-playground/validator/v10 v10.4.1
	github.com/go-resty/resty/v2 v2.5.0
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/jarcoal/httpmock v1.0.8
	github.com/nats-io/nats-server/v2 v2.1.7
	github.com/nats-io/nats.go v1.10.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/alertmanager v0.21.0
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/sclevine/agouti v0.0.0-20190613051229-00c1187c74ad
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/tebeka/selenium v0.9.9
	github.com/weaveworks/cluster-api-provider-existinginfra v0.2.5
	github.com/weaveworks/libgitops v0.0.3
	github.com/weaveworks/pctl v0.8.0
	github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server v0.0.0-00010101000000-000000000000
	github.com/weaveworks/weave-gitops-enterprise/common v0.0.0
	github.com/weaveworks/wksctl v0.10.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gorm.io/datatypes v1.0.0
	gorm.io/driver/sqlite v1.1.4 // indirect
	gorm.io/gorm v1.21.11
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/cli-runtime v0.21.2
	k8s.io/client-go v0.22.1
	k8s.io/kubernetes v1.21.1 // indirect
	sigs.k8s.io/cluster-api v0.3.16
	sigs.k8s.io/controller-runtime v0.9.6
	sigs.k8s.io/kustomize/kyaml v0.11.1
)

require (
	cloud.google.com/go v0.81.0 // indirect
	code.gitea.io/sdk/gitea v0.14.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.18 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.13 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20210428141323-04723f9f07d7 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535 // indirect
	github.com/aws/eks-distro-build-tooling/release v0.0.0-20201211225747-07f05a470de8 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitnami-labs/sealed-secrets v0.12.5 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/bluekeyes/go-gitdiff v0.4.0 // indirect
	github.com/cavaliercoder/go-rpm v0.0.0-20200122174316-8cb9fd9c31a8 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/chanwit/plandiff v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/drone/envsubst v1.0.3-0.20200709223903-efdb65b94e5a // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/fluxcd/go-git-providers v0.2.1-0.20210908163615-833963032787 // indirect
	github.com/fluxcd/helm-controller/api v0.11.2 // indirect
	github.com/fluxcd/kustomize-controller/api v0.14.0 // indirect
	github.com/fluxcd/pkg/apis/kustomize v0.2.0 // indirect
	github.com/fluxcd/pkg/apis/meta v0.10.1 // indirect
	github.com/fluxcd/pkg/runtime v0.12.1 // indirect
	github.com/fluxcd/source-controller/api v0.15.4 // indirect
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-errors/errors v1.4.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/go-git/go-git/v5 v5.4.2 // indirect
	github.com/go-logr/logr v1.1.0 // indirect
	github.com/go-logr/zapr v1.1.0 // indirect
	github.com/go-openapi/analysis v0.19.10 // indirect
	github.com/go-openapi/errors v0.19.4 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/loads v0.19.5 // indirect
	github.com/go-openapi/runtime v0.19.15 // indirect
	github.com/go-openapi/spec v0.19.8 // indirect
	github.com/go-openapi/strfmt v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/go-openapi/validate v0.19.8 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-github/v32 v32.1.0 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.5.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.8 // indirect
	github.com/hashicorp/go-version v1.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.7.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.0.5 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.5.0 // indirect
	github.com/jackc/pgx/v4 v4.9.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jenkins-x/go-scm v1.10.10 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lightstep/tracecontext.go v0.0.0-20181129014701-1757c391b1ac // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mattn/go-sqlite3 v1.14.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/copystructure v1.1.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/mkmik/multierror v0.3.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/nats-io/jwt v0.3.2 // indirect
	github.com/nats-io/nkeys v0.1.4 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/otiai10/copy v1.6.0 // indirect
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.29.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shurcooL/githubv4 v0.0.0-20190718010115-4ba037080260 // indirect
	github.com/shurcooL/graphql v0.0.0-20181231061246-d48a9a75455f // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.8.1 // indirect
	github.com/src-d/gcfg v1.4.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/weaveworks/go-checkpoint v0.0.0-20170503165305-ebbb8b0518ab // indirect
	github.com/weaveworks/profiles v0.2.0 // indirect
	github.com/weaveworks/weave-gitops v0.2.6-0.20210915135659-69953b81831d // indirect
	github.com/weaveworks/weave-gitops-enterprise-credentials v0.0.1 // indirect
	github.com/xanzy/go-gitlab v0.43.0 // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	go.mongodb.org/mongo-driver v1.3.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.19.0 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5 // indirect
	golang.org/x/oauth2 v0.0.0-20210615190721-d04028783cf1 // indirect
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0 // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210617175327-b9e0b3197ced // indirect
	google.golang.org/grpc v1.39.1 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/oleiade/reflections.v1 v1.0.0 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/postgres v1.0.5 // indirect
	k8s.io/apiextensions-apiserver v0.21.3 // indirect
	k8s.io/cluster-bootstrap v0.20.2 // indirect
	k8s.io/component-base v0.22.1 // indirect
	k8s.io/klog/v2 v2.10.0 // indirect
	k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e // indirect
	k8s.io/kube-proxy v0.0.0 // indirect
	k8s.io/utils v0.0.0-20210722164352-7f3ee0f31471 // indirect
	knative.dev/pkg v0.0.0-20210412173742-b51994e3b312 // indirect
	sigs.k8s.io/kustomize/api v0.9.0 // indirect
	sigs.k8s.io/kustomize/kstatus v0.0.2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
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

replace github.com/go-logr/logr v1.1.0 => github.com/go-logr/logr v0.4.0

replace github.com/go-logr/zapr v1.1.0 => github.com/go-logr/zapr v0.4.0
