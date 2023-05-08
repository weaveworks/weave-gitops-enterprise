# Weave-Gitops-Enterprise Release 0.23.0 notes

Release data: 2023-05-11

## Compatibility

WGE 0.23.0 requires at least Flux `v2.0.0-rc.1`.

_Note: Flux v2.0.0-rc.1 has the same kubernetes compatibility as Flux 0.41.2_

If you using a hosted flux version, please check with your provider if they support the new version before upgrading to 0.23.0. Known hosted flux providers:

- EKS Anywhere
- [Azure AKS Flux-Gitops extension](https://learn.microsoft.com/en-us/azure/azure-arc/kubernetes/extensions-release#flux-gitops)

As of writing they do not yet support the new version so please wait before upgrading to WGE 0.23.0.

## Migrating your EE installation to Flux GA

Below we'll take you through the multiple steps required to migrate to 0.23.0.

After each step the cluster will be in a working state, so you can take your time to complete the migration.

1. Upgrade to WGE 0.22.0
2. Upgrade to Flux v2.0.0-rc.1 on your leaf clusters and management clusters
3. Upgrade templates, gitopssets and cluster bootstrap configs
4. Upgrade to Flux v2.0.0-rc.1 in `ClusterBootstrapConfig`s
5. Upgrade to WGE 0.23.0

### 1. Upgrade to WGE 0.22.0

WGE 0.22.0 is compatible with Flux v2.0.0-rc.1 (except for gitopssets) and will make sure your cluster is in a working state before upgrading to WGE 0.23.0.

If you are using gitopssets we can upgrade that component to gitopssets v0.10.0 for flux v2.0.0-rc.1 compatibility. Update the Weave Gitops Enterprise HelmRelease values to use the new version.

```yaml
gitopssets-controller:
  controllerManager:
    manager:
      image:
        tag: v0.10.0
```

### 2. Upgrade to Flux v2.0.0-rc.1 on your leaf clusters and management clusters

Follow the upgrade instuctions from the [Flux v2.0.0-rc.1 release notes](https://github.com/fluxcd/flux2/releases/tag/v2.0.0-rc.1)

### 3. Upgrade templates, gitopssets and cluster bootstrap configs

#### `GitopsTemplate` and `CAPITemplate`

Update `GitRepository` and `Kustomization` CRs in the `spec.resourcetemplates` to `v1` as described in the flux upgrade instructions.

#### `GitopsSets`

Update `GitRepository` and `Kustomization` CRs in the `spec.template` of your `GitopsSet` resources to `v1` as described in the flux upgrade instructions.

#### `ClusterBootstrapConfig`

`ClusterBootstrapConfig` will most often contain an invocation of `flux bootstrap`, make sure the image is using `v2`

```patch
diff --git a/tools/dev-resources/user-guide/cluster-bootstrap-config.yaml b/tools/dev-resources/user-guide/cluster-bootstrap-config.yaml
index bd41ec036..1b21df860 100644
--- a/tools/dev-resources/user-guide/cluster-bootstrap-config.yaml
+++ b/tools/dev-resources/user-guide/cluster-bootstrap-config.yaml
@@ -1,34 +1,34 @@
 apiVersion: capi.weave.works/v1alpha1
 kind: ClusterBootstrapConfig
 metadata:
   name: capi-gitops
   namespace: default
 spec:
   clusterSelector:
     matchLabels:
       weave.works/capi: bootstrap
   jobTemplate:
     generateName: "run-gitops-{{ .ObjectMeta.Name }}"
     spec:
       containers:
-        - image: ghcr.io/fluxcd/flux-cli:v0.34.0
+        - image: ghcr.io/fluxcd/flux-cli:v2
           name: flux-bootstrap
           resources: {}
           volumeMounts:
             - name: kubeconfig
               mountPath: "/etc/gitops"
               readOnly: true
           args:
             [
               "bootstrap",
               "github",
               "--kubeconfig=/etc/gitops/value",
               "--owner=$(GITHUB_USER)",
               "--repository=$(GITHUB_REPO)",
               "--path=./clusters/{{ .ObjectMeta.Namespace }}/{{ .ObjectMeta.Name }}",
             ]
           envFrom:
             - secretRef:
                 name: my-pat
       restartPolicy: Never
       volumes:
```

### 4. Upgrade to WGE 0.23.0

Upgrade the Weave Gitops Enterprise HelmRelease values to use the new version.

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: weave-gitops-enterprise
  namespace: flux-system
spec:
  chart:
    spec:
      version: 0.22.0
```

## WGE 0.23.0

WGE 0.23.0's features will now generate `v1` Kustomizations:

- Add app
- Common bases Kustomization for `GitopsTemplate` and `CAPITemplate`s.
