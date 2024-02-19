module github.com/weaveworks/weave-gitops-enterprise

go 1.20

require (
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/go-logr/logr v1.3.0
	github.com/google/go-cmp v0.6.0
	github.com/google/uuid v1.3.1
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/onsi/gomega v1.30.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.16.0
	github.com/sclevine/agouti v3.0.0+incompatible
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.7.0
	github.com/stretchr/testify v1.8.4
	github.com/weaveworks/weave-gitops v0.38.1-0.20231228113211-a38fbeca6a75
	github.com/weaveworks/weave-gitops-enterprise/common v0.0.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.27.7
	k8s.io/apimachinery v0.27.7
	k8s.io/cli-runtime v0.27.2
	k8s.io/client-go v1.5.2
	sigs.k8s.io/controller-runtime v0.15.2
)

require (
	filippo.io/age v1.1.1
	github.com/NYTimes/gziphandler v1.1.1
	github.com/ProtonMail/gopenpgp/v2 v2.6.0
	github.com/alecthomas/assert v0.0.0-20170929043011-405dbfeb8e38
	github.com/alexedwards/scs/v2 v2.5.1
	github.com/aws/aws-sdk-go-v2 v1.16.16
	github.com/aws/aws-sdk-go-v2/config v1.17.8
	github.com/aws/aws-sdk-go-v2/service/pricing v1.17.1
	github.com/external-secrets/external-secrets v0.9.3
	github.com/fluxcd/flagger v1.30.0
	github.com/fluxcd/go-git-providers v0.16.0
	github.com/fluxcd/helm-controller/api v0.35.0
	github.com/fluxcd/kustomize-controller/api v1.0.0
	github.com/fluxcd/pkg/apis/meta v1.1.2
	github.com/fluxcd/pkg/runtime v0.42.0
	github.com/fluxcd/pkg/untar v0.2.0
	github.com/fluxcd/pkg/version v0.2.1
	github.com/fluxcd/source-controller/api v1.0.0
	github.com/go-resty/resty/v2 v2.7.0
	github.com/golang/protobuf v1.5.3
	github.com/google/go-github/v32 v32.1.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.10.2
	github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts v1.1.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jarcoal/httpmock v1.0.8
	github.com/jenkins-x/go-scm v1.14.14
	github.com/loft-sh/vcluster v0.12.0
	github.com/mattn/go-sqlite3 v1.14.16
	github.com/maxbrunsfeld/counterfeiter/v6 v6.6.2
	github.com/microsoft/azure-devops-go-api/azuredevops v1.0.0-b5
	github.com/mkmik/multierror v0.3.0
	github.com/onsi/ginkgo/v2 v2.13.0
	github.com/slok/go-http-metrics v0.10.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.16.0
	github.com/tonglil/buflogr v1.0.1
	github.com/weaveworks/cluster-controller v1.5.5
	github.com/weaveworks/cluster-reflector-controller v0.0.9
	github.com/weaveworks/gitopssets-controller v0.16.5
	github.com/weaveworks/policy-agent/api v1.0.5
	github.com/weaveworks/progressive-delivery v0.0.0-20230421131659-61a8aadf8aac
	github.com/weaveworks/templates-controller v0.2.0
	github.com/xanzy/go-gitlab v0.90.0
	go.mozilla.org/sops/v3 v3.7.3
	golang.org/x/crypto v0.17.0
	golang.org/x/exp v0.0.0-20230811145659-89c5cff77bcb
	golang.org/x/oauth2 v0.13.0
	google.golang.org/genproto/googleapis/api v0.0.0-20230822172742-b8732ec3820d
	google.golang.org/grpc v1.59.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.31.0
	helm.sh/helm/v3 v3.11.3
	k8s.io/apiextensions-apiserver v0.27.4
	k8s.io/kubernetes v1.26.3
	sigs.k8s.io/cluster-api v1.5.2
	sigs.k8s.io/kustomize/kyaml v0.14.1
	sigs.k8s.io/yaml v1.4.0
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/fluxcd/pkg/tar v0.4.0 // indirect
	github.com/gitops-tools/pkg v0.1.0 // indirect
	github.com/google/go-containerregistry v0.12.0 // indirect
)

require (
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230106234847-43070de90fa1 // indirect
	github.com/RoaringBitmap/roaring v0.9.4 // indirect
	github.com/bits-and-blooms/bitset v1.2.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/blevesearch/bleve_index_api v1.0.5 // indirect
	github.com/blevesearch/geo v0.1.17 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/gtreap v0.1.1 // indirect
	github.com/blevesearch/mmap-go v1.0.4 // indirect
	github.com/blevesearch/scorch_segment_api/v2 v2.1.4 // indirect
	github.com/blevesearch/segment v0.9.1 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/upsidedown_store_api v1.0.2 // indirect
	github.com/blevesearch/vellum v1.0.9 // indirect
	github.com/blevesearch/zapx/v11 v11.3.7 // indirect
	github.com/blevesearch/zapx/v12 v12.3.7 // indirect
	github.com/blevesearch/zapx/v13 v13.3.7 // indirect
	github.com/blevesearch/zapx/v14 v14.3.7 // indirect
	github.com/blevesearch/zapx/v15 v15.3.9 // indirect
	github.com/go-jose/go-jose/v3 v3.0.1 // indirect
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // indirect
	github.com/google/go-github/v52 v52.0.0 // indirect
	github.com/google/s2a-go v0.1.5 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	go.etcd.io/bbolt v1.3.7 // indirect
	go.opentelemetry.io/otel/metric v1.20.0 // indirect
	google.golang.org/api v0.136.0 // indirect
	google.golang.org/genproto v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	gopkg.in/urfave/cli.v1 v1.20.0 // indirect
	k8s.io/component-helpers v0.27.2 // indirect
)

require (
	cloud.google.com/go/compute v1.23.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	code.gitea.io/sdk/gitea v0.14.0 // indirect
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.29 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.23 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.12 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.6 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/ProtonMail/go-mime v0.0.0-20221031134845-8fd9bc37cf08 // indirect
	github.com/alecthomas/colour v0.0.0-20160524082231-60882d9e2721 // indirect
	github.com/alecthomas/repr v0.0.0-20180818092828-117648cd9897 // indirect
	github.com/aws/aws-sdk-go v1.44.322 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.12.21 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.24 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.13.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.19 // indirect
	github.com/aws/smithy-go v1.13.3 // indirect
	github.com/bluekeyes/go-gitdiff v0.7.1 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fluxcd/image-automation-controller/api v0.33.1 // indirect
	github.com/fluxcd/image-reflector-controller/api v0.27.2 // indirect
	github.com/fluxcd/notification-controller/api v1.0.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/pprof v0.0.0-20230811205829-9131a7e9cc17 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.5 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/goware/prefixer v0.0.0-20160118172347-395022866408 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.7 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/vault/api v1.9.2 // indirect
	github.com/howeyc/gopass v0.0.0-20210920133722-c8aef6fb66ef // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/shurcooL/githubv4 v0.0.0-20190718010115-4ba037080260 // indirect
	github.com/shurcooL/graphql v0.0.0-20181231061246-d48a9a75455f // indirect
	github.com/skeema/knownhosts v1.2.1 // indirect
	go.mozilla.org/gopgagent v0.0.0-20170926210634-4d7ea76ff71a // indirect
	golang.org/x/tools v0.14.0 // indirect
	gorm.io/gorm v1.24.0
)

require (
	github.com/bufbuild/connect-go v0.2.0 // indirect
	github.com/chzyer/readline v1.5.1
	github.com/containerd/typeurl v1.0.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/go-chi/chi/v5 v5.0.7 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gofrs/uuid v4.2.0+incompatible // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/jdxcode/netrc v0.0.0-20210204082910-926c7f70242a // indirect
	github.com/jhump/protocompile v0.0.0-20220216033700-d705409f108f // indirect
	github.com/jhump/protoreflect v1.12.1-0.20220721211354-060cc04fc18b // indirect
	github.com/klauspost/cpuid v1.3.1 // indirect
	github.com/klauspost/pgzip v1.2.5 // indirect
	github.com/manifoldco/promptui v0.9.0
	github.com/minio/md5-simd v1.1.0 // indirect
	github.com/minio/minio-go/v7 v7.0.31 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/moby/buildkit v0.10.3 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pkg/profile v1.6.0 // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/rs/xid v1.4.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/weaveworks/pipeline-controller/api v0.0.0-20230228164807-3af8aa2ecc3d
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.46.0 // indirect
	go.opentelemetry.io/otel v1.20.0 // indirect
	go.opentelemetry.io/otel/trace v1.20.0 // indirect
	golang.org/x/mod v0.13.0 // indirect
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/Masterminds/squirrel v1.5.3 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230828082145-3c4c8a2d2371 // indirect
	github.com/alecthomas/chroma v0.9.2 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/blevesearch/bleve/v2 v2.3.7
	github.com/bufbuild/buf v1.7.0
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/cheshir/ttlcache v1.0.1-0.20220504185148-8ceeff21b789
	github.com/clbanning/mxj/v2 v2.7.0 // indirect
	github.com/containerd/containerd v1.7.0 // indirect
	github.com/coreos/go-oidc/v3 v3.4.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/docker/cli v20.10.21+incompatible // indirect
	github.com/docker/docker v20.10.24+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.7.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/drone/envsubst/v2 v2.0.0-20210730161058-179042472c46 // indirect
	github.com/emicklei/go-restful/v3 v3.10.2 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/fluxcd/pkg/apis/acl v0.1.0 // indirect
	github.com/fluxcd/pkg/apis/kustomize v1.1.1 // indirect
	github.com/fluxcd/pkg/ssa v0.27.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-asset/generics v0.0.0-20220317100214-d5f632c68060 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.5.0 // indirect
	github.com/go-git/go-git/v5 v5.11.0
	github.com/go-gorp/gorp/v3 v3.0.5 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.20.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/glog v1.1.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.5 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-5 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/iancoleman/strcase v0.2.0 // indirect
	github.com/imdario/mergo v0.3.16
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2.0.20221005185240-3a7f492d3f1b // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/rivo/uniseg v0.4.2 // indirect
	github.com/rubenv/sql-migrate v1.3.1 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/sethvargo/go-limiter v0.7.2 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/tomwright/dasel v1.22.1 // indirect
	github.com/weaveworks/tf-controller/api v0.0.0-20231101110059-994a65055198
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.starlark.net v0.0.0-20230302034142-4b1e35fe2254 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.26.0
	golang.org/x/net v0.19.0
	golang.org/x/sync v0.4.0
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/term v0.15.0 // indirect
	golang.org/x/text v0.14.0
	golang.org/x/time v0.3.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/sqlite v1.4.4
	k8s.io/apiserver v0.27.2
	k8s.io/component-base v0.27.4 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f // indirect
	k8s.io/kubectl v0.27.2
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b
	oras.land/oras-go v1.2.2 // indirect
	sigs.k8s.io/cli-utils v0.35.0 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/kustomize/api v0.13.2 // indirect
	sigs.k8s.io/kustomize/kstatus v0.0.2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

replace (
	// TODO: why do we need to replace this?
	github.com/appscode/jsonpatch => gomodules.xyz/jsonpatch/v2 v2.0.0

	// replace circl to 1.3.3
	// https://github.com/advisories/GHSA-2q89-485c-9j2x
	github.com/cloudflare/circl => github.com/cloudflare/circl v1.3.3

	// Replace digest lib to master to gather access to BLAKE3.
	// xref: https://github.com/opencontainers/go-digest/pull/66
	github.com/opencontainers/go-digest => github.com/opencontainers/go-digest v1.0.1-0.20220411205349-bde1400a84be

	// un-comment for local dev
	// github.com/weaveworks/weave-gitops => ../weave-gitops

	github.com/weaveworks/weave-gitops-enterprise/common => ./common

	//
	// As we import k8s.io/kubernetes, we need to replace the following modules
	// with the same version as k8s.io/kubernetes. In short, you are not supposed
	// to use k8s.io/kubernetes as a module (https://github.com/kubernetes/kubernetes/issues/81878#issuecomment-696689706).
	// The module declares a lot of internal replacements, which we have to re-replace here.
	//
	// We need some things in there like the RBAC APIs which are not exported elsewhere.
	//
	// If you add a new k8s.io/* dependency, add it here as well.
	// If you forget to do this, you will see something like this next time you run `go get`:
	//
	// $ go get github.com/weaveworks/weave-gitops@v0.25.1-rc.1
	// go: downgraded k8s.io/kubernetes v1.26.3 => v1.15.0-alpha.0
	//
	// And the project won't build anymore.
	//
	k8s.io/api => k8s.io/api v0.27.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.26.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.27.7
	k8s.io/apiserver => k8s.io/apiserver v0.26.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.27.7
	k8s.io/client-go => k8s.io/client-go v0.27.7
	k8s.io/component-base => k8s.io/component-base v0.26.3
	k8s.io/component-helpers => k8s.io/component-helpers v0.26.3
	k8s.io/kubectl => k8s.io/kubectl v0.26.3
)
