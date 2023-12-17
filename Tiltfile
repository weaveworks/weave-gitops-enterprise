load('ext://restart_process', 'docker_build_with_restart')
load('ext://helm_remote', 'helm_remote')

if os.path.exists('Tiltfile.local'):
   include('Tiltfile.local')

# Don't replace "dev" (password) in the logs with "[redacted]"
secret_settings(disable_scrub=True)

if os.getenv('MANUAL_MODE'):
   trigger_mode(TRIGGER_MODE_MANUAL)

# Download chart deps on first run. This command is slow, so you'd have to
# re-run it yourself if you upgrade the chart or the chart changes at all.
# By declaring a local_resource with TRIGGER_MODE_MANUAL we can have it run
# only once on `tilt up`
# https://docs.tilt.dev/local_resource.html#file-dependencies
local_resource("helm-dep-update", "helm dep update charts/mccp", trigger_mode=TRIGGER_MODE_MANUAL, auto_init=True)

# This is needed for javascript access
if not os.getenv('GITHUB_TOKEN'):
   fail("You need to set GITHUB_TOKEN in your terminal before running this")

# --- what to edit
# e.g.
# TO_EDIT="clusters-controller" tilt up

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

gitopssets_controller_labels = ["remote-images"]
if 'gitopssets-controller' in to_edit:
   if not os.path.exists("../gitopssets-controller"):
      fail("You need to git clone https://github.com/weaveworks/gitopssets-controller to a directory next to this")
   docker_build('ghcr.io/weaveworks/gitopssets-controller', '../gitopssets-controller/')
   gitopssets_controller_labels = ["local"]

cluster_reflector_controller_labels = ["remote-images"]
if 'cluster-reflector-controller' in to_edit:
   if not os.path.exists("../cluster-reflector-controller"):
      fail("You need to git clone https://github.com/weaveworks/cluster-reflector-controller to a directory next to this")
   docker_build('ghcr.io/weaveworks/cluster-reflector-controller', '../cluster-reflector-controller/')
   cluster_reflector_controller_labels = ["local"]

# --- rename chart resources to human readable 

k8s_resource('chart-mccp-cluster-service', new_name='cluster-service', labels=["local"], port_forwards='8000')
k8s_resource('chart-pipeline-controller', new_name='pipeline-controller', labels=["remote-images"])
k8s_resource('cluster-bootstrap-controller-manager', new_name='cluster-bootstrap-controller', labels=cluster_bootstrap_controller_labels)
k8s_resource('cluster-controller-manager', new_name='cluster-controller', labels=cluster_controller_labels)
k8s_resource('gitopssets-controller-manager', new_name='gitopssets-controller', labels=gitopssets_controller_labels)
k8s_resource('cluster-reflector-controller-manager', labels=cluster_reflector_controller_labels)
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
   print("Skipping UI build")
   dockerfile = """
   FROM alpine:3.18
   """
   docker_build("weaveworks/weave-gitops-enterprise-ui-server", "", dockerfile_contents=dockerfile, ignore=["ui", "build", ".parcel-cache"])

elif native_build:
   # Build UI locally

   local_resource(
      'ui-native-build',
      'make ui-build',
      deps=[
         './ui/src',
      ],
      labels=['local'],
   )
   docker_build(
      'weaveworks/weave-gitops-enterprise-ui-server',
      '.',
      dockerfile="ui/dev.dockerfile",
      build_args={'GITHUB_TOKEN': os.getenv('GITHUB_TOKEN')},
   )

else:
   # Build UI in container (default)

   docker_build(
      'weaveworks/weave-gitops-enterprise-ui-server',
      '.',
      dockerfile='ui/Dockerfile',
      build_args={'GITHUB_TOKEN': os.getenv('GITHUB_TOKEN')},
   )

# --- clusters-service

if native_build:
   # Build locally (usually slower under MacOS than build in container)

   local_resource(
      'compile',
      'make build-linux',
      deps=[
         '../weave-gitops/core',
         './cmd/clusters-service',
         './pkg'
      ],
      ignore=[
         './cmd/clusters-service/bin',
         '.parcel-cache',
      ],
      dir='cmd/clusters-service',
      labels=["local"]
   )

   docker_build_with_restart(
      'weaveworks/weave-gitops-enterprise-clusters-service',
      '.',
      dockerfile="cmd/clusters-service/dev.dockerfile",
      entrypoint='/app/clusters-service --log-level=debug',
      build_args={'GITHUB_BUILD_TOKEN': os.getenv('GITHUB_TOKEN'), 'image_tag': 'tilt'},
      live_update=[
         sync('cmd/clusters-service/bin', '/app'),
      ],
      ignore=[
         'cmd/clusters-service/clusters-service',
         './build',
         '.parcel-cache',
      ],
   )
else:
   # Build in container (default)

   docker_build(
      'weaveworks/weave-gitops-enterprise-clusters-service',
      '.',
      ignore=["ui", "build", ".parcel-cache"],
      dockerfile='cmd/clusters-service/Dockerfile',
      build_args={'GITHUB_BUILD_TOKEN': os.getenv('GITHUB_TOKEN'),'image_tag': 'tilt'},
      entrypoint= ["/clusters-service", "--log-level=debug"],
   )
