load('ext://restart_process', 'docker_build_with_restart')
load('ext://helm_remote', 'helm_remote')

if os.getenv('MANUAL_MODE'):
   trigger_mode(TRIGGER_MODE_MANUAL)

if not os.path.exists("./charts/mccp/charts"):
   # Download chart deps on first run. This command is slow, so you'd have to
   # re-run it yourself if you upgrade the chart
   local("helm dep update charts/mccp")

if not os.path.exists("../cluster-bootstrap-controller"):
   fail("You need to git clone https://github.com/weaveworks/cluster-bootstrap-controller to a directory next to this")

if not os.path.exists("../cluster-controller"):
   fail("You need to git clone https://github.com/weaveworks/cluster-controller to a directory next to this")


# This is needed for javascript access
if not os.getenv('GITHUB_TOKEN'):
   fail("You need to set GITHUB_TOKEN in your terminal before running this")

# Install resources I couldn't find elsewhere
k8s_yaml(listdir('tools/dev-resources/', recursive=True))

k8s_yaml('test/utils/scripts/entitlement-secret.yaml')

helm_values = ['tools/dev-values.yaml']
if os.path.exists('tools/dev-values-local.yaml'):
   helm_values.append('tools/dev-values-local.yaml')

k8s_yaml(helm(
    "charts/mccp",
    namespace='flux-system',
    values=helm_values,
))

k8s_yaml(kustomize('../cluster-controller/config/crd'))
k8s_yaml(kustomize('../cluster-bootstrap-controller/config/crd'))

docker_build('weaveworks/cluster-controller', '../cluster-controller/')
docker_build('weaveworks/cluster-bootstrap-controller', '../cluster-bootstrap-controller/',
   build_args={'GITHUB_BUILD_USERNAME': 'wge-build-bot', 'GITHUB_BUILD_TOKEN': os.getenv('GITHUB_TOKEN')}
)

helm_remote('tf-controller',
            repo_url='https://weaveworks.github.io/tf-controller',
            namespace='flux-system')

# Note for MacOS users:
# for this to work you need to run:
#   brew install FiloSottile/musl-cross/musl-cross
# https://github.com/mattn/go-sqlite3#cross-compiling-from-mac-osx
native_build = os.getenv('NATIVE_BUILD', False)
skip_ui = os.getenv("SKIP_UI_BUILD", False)
if native_build:
   local_resource(
      'clusters-service',
      'make build-linux',
      deps=[
         './cmd/clusters-service',
         './pkg'
      ],
      ignore=[
         './cmd/clusters-service/bin'
      ],
      dir='cmd/clusters-service',
   )

   if not skip_ui:
      local_resource(
         'ui',
         'make build',
         deps=[
            './ui-cra/src',
         ],
         dir='ui-cra',
      )

   docker_build_with_restart(
      'weaveworks/weave-gitops-enterprise-clusters-service',
      '.',
      dockerfile="cmd/clusters-service/dev.dockerfile",
      entrypoint='/app/clusters-service',
      build_args={'GITHUB_BUILD_TOKEN': os.getenv('GITHUB_TOKEN'), 'image_tag': 'tilt'},
      live_update=[
         sync('cmd/clusters-service/bin', '/app'),
      ],
      ignore=[
         'cmd/clusters-service/clusters-service'
      ]
   )

   docker_build(
      'weaveworks/weave-gitops-enterprise-ui-server',
      'ui-cra',
      dockerfile="ui-cra/dev.dockerfile",
      build_args={'GITHUB_TOKEN': os.getenv('GITHUB_TOKEN')},
   )
else:
   docker_build(
      'weaveworks/weave-gitops-enterprise-clusters-service',
      '.',
      dockerfile='cmd/clusters-service/Dockerfile',
      build_args={'GITHUB_BUILD_TOKEN': os.getenv('GITHUB_TOKEN'),'image_tag': 'tilt'},
      entrypoint= ["/sbin/tini", "--", "clusters-service", "--dev-mode"]
   )
   docker_build(
      'weaveworks/weave-gitops-enterprise-ui-server',
      'ui-cra',
      build_args={'GITHUB_TOKEN': os.getenv('GITHUB_TOKEN')}
   )

k8s_resource('chart-mccp-cluster-service', port_forwards='8000')

secret_settings(disable_scrub=True)
