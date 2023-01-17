load('ext://restart_process', 'docker_build_with_restart')
load('ext://helm_remote', 'helm_remote')

if os.path.exists('Tiltfile.local'):
   include('Tiltfile.local')

# Don't replace "dev" (password) in the logs with "[redacted]"
secret_settings(disable_scrub=True)

if os.getenv('MANUAL_MODE'):
   trigger_mode(TRIGGER_MODE_MANUAL)

if not os.path.exists("./charts/mccp/charts"):
   # Download chart deps on first run. This command is slow, so you'd have to
   # re-run it yourself if you upgrade the chart
   local("helm dep update charts/mccp")

# This is needed for javascript access
if not os.getenv('GITHUB_TOKEN'):
   fail("You need to set GITHUB_TOKEN in your terminal before running this")

# --- what to edit
# e.g.
# TO_EDIT="templates-controller,clusters-controller" tilt up


to_edit = os.getenv('TO_EDIT', '').split(",")

cluster_bootstrap_controller_labels = ["remote-images"]
if 'cluster-bootstrap-controller' in to_edit:
   if not os.path.exists("../cluster-bootstrap-controller"):
      fail("You need to git clone https://github.com/weaveworks/cluster-bootstrap-controller to a directory next to this")
   docker_build('weaveworks/cluster-bootstrap-controller', '../cluster-bootstrap-controller/',
      build_args={'GITHUB_BUILD_USERNAME': 'wge-build-bot', 'GITHUB_BUILD_TOKEN': os.getenv('GITHUB_TOKEN')}
   )
   cluster_bootstrap_controller_labels = ["local"]

cluster_controller_labels = ["remote-images"]
if 'cluster-controller' in to_edit:
   if not os.path.exists("../cluster-controller"):
      fail("You need to git clone https://github.com/weaveworks/cluster-controller to a directory next to this")
   docker_build('weaveworks/cluster-controller', '../cluster-controller/')
   cluster_bootstrap_controller_labels = ["local"]

templates_controller_labels = ["remote-images"]
if 'templates-controller' in to_edit:
   if not os.path.exists("../templates-controller"):
      fail("You need to git clone https://github.com/weaveworks/templates-controller to a directory next to this")
   docker_build('ghcr.io/weaveworks/templates-controller', '../templates-controller/')
   templates_controller_labels = ["local"]

gitopssets_controller_labels = ["remote-images"]
if 'gitopssets-controller' in to_edit:
   if not os.path.exists("../gitopssets-controller"):
      fail("You need to git clone https://github.com/weaveworks/gitopssets-controller to a directory next to this")
   docker_build('ghcr.io/weaveworks/gitopssets-controller', '../gitopssets-controller/')
   templates_controller_labels = ["local"]

# --- rename chart resources to human readable 

k8s_resource('chart-mccp-cluster-service', new_name='cluster-service', labels=["local"], port_forwards='8000')
k8s_resource('chart-pipeline-controller', new_name='pipeline-controller', labels=["remote-images"])
k8s_resource('chart-mccp-cluster-bootstrap-controller', new_name='cluster-bootstrap-controller', labels=cluster_bootstrap_controller_labels)
k8s_resource('chart-cluster-controller', new_name='cluster-controller', labels=cluster_controller_labels)
k8s_resource('templates-controller-controller-manager', new_name='templates-controller', labels=templates_controller_labels)
k8s_resource('gitopssets-controller-manager', new_name='gitopssets-controller', labels=gitopssets_controller_labels)
k8s_resource('policy-agent', labels=["remote-images"])

# Install resources I couldn't find elsewhere
k8s_yaml(listdir('tools/dev-resources/', recursive=True))

k8s_yaml('test/utils/data/entitlement/entitlement-secret.yaml')

helm_values = ['tools/dev-values.yaml']
if os.path.exists('tools/dev-values-local.yaml'):
   helm_values.append('tools/dev-values-local.yaml')

k8s_yaml(helm(
   "charts/mccp",
   namespace='flux-system',
   values=helm_values,
))

# --- tf-controller
#
# install the external tf-controller chart too
helm_remote('tf-controller',
            repo_url='https://weaveworks.github.io/tf-controller',
            namespace='flux-system')
k8s_resource('tf-controller', labels=["remote-images"])

# Note for MacOS users:
# Not recommended, it will be slower than build in container and
# for this to work you need to run:
#   brew install FiloSottile/musl-cross/musl-cross
# https://github.com/mattn/go-sqlite3#cross-compiling-from-mac-osx
native_build = os.getenv('NATIVE_BUILD', False)
skip_ui = os.getenv("SKIP_UI_BUILD", False)

# --- ui

if skip_ui:
   dockerfile = """
   FROM alpine:3.13
   """
   docker_build("weaveworks/weave-gitops-enterprise-ui-server", "", dockerfile_contents=dockerfile)

elif native_build:
   # Build UI locally

   local_resource(
      'ui-native-build',
      'make build',
      deps=[
         './ui-cra/src',
      ],
      dir='ui-cra',
      labels=['local'],
   )
   docker_build(
      'weaveworks/weave-gitops-enterprise-ui-server',
      'ui-cra',
      dockerfile="ui-cra/dev.dockerfile",
      build_args={'GITHUB_TOKEN': os.getenv('GITHUB_TOKEN')},
   )

else:
   # Build UI in container (default)

   docker_build(
      'weaveworks/weave-gitops-enterprise-ui-server',
      'ui-cra',
      build_args={'GITHUB_TOKEN': os.getenv('GITHUB_TOKEN')},
   )

# --- clusters-service

if native_build:
   # Build locally (usually slower under MacOS than build in container)

   local_resource(
      'clusters-service-native-build',
      'make build-linux',
      deps=[
         './cmd/clusters-service',
         './pkg'
      ],
      ignore=[
         './cmd/clusters-service/bin'
      ],
      dir='cmd/clusters-service',
      labels=["local"]
   )

   docker_build_with_restart(
      'weaveworks/weave-gitops-enterprise-clusters-service',
      '.',
      dockerfile="cmd/clusters-service/dev.dockerfile",
      entrypoint='/app/clusters-service --dev-mode',
      build_args={'GITHUB_BUILD_TOKEN': os.getenv('GITHUB_TOKEN'), 'image_tag': 'tilt'},
      live_update=[
         sync('cmd/clusters-service/bin', '/app'),
      ],
      ignore=[
         'cmd/clusters-service/clusters-service'
      ],
   )
else:
   # Build in container (default)

   docker_build(
      'weaveworks/weave-gitops-enterprise-clusters-service',
      '.',
      ignore=["ui-cra"],
      dockerfile='cmd/clusters-service/Dockerfile',
      build_args={'GITHUB_BUILD_TOKEN': os.getenv('GITHUB_TOKEN'),'image_tag': 'tilt'},
      entrypoint= ["/sbin/tini", "--", "clusters-service", "--dev-mode"]
   )
