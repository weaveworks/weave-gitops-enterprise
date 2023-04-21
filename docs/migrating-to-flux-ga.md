## Migrating your EE installation to Flux GA

### `GitopsTemplate` and `CAPITemplate`

- Update all Flux CRs in the `spec.resourcetemplates`
- Update `spec.charts.items[].template.content` to make sure its using v1 fields

?? After updating the templates you can _edit_ them via the UI to re-render and update the generated flux resources ??

### `GitopsSets`

- Update all Flux CRs in the `spec.template` of your `GitopsSet` resources.

### `Pipeline`

- Update all the `spec.appRef.apiVersion` in your `Pipeline` resources.

```patch
diff --git a/tools/dev-resources/pipelines/github/pipeline.yaml b/tools/dev-resources/pipelines/github/pipeline.yaml
index b5eb66b0b..ca321a30c 100644
--- a/tools/dev-resources/pipelines/github/pipeline.yaml
+++ b/tools/dev-resources/pipelines/github/pipeline.yaml
@@ -1,18 +1,18 @@
 apiVersion: pipelines.weave.works/v1alpha1
 kind: Pipeline
 metadata:
   name: podinfo-github
   namespace: flux-system
 spec:
   appRef:
-    apiVersion: helm.toolkit.fluxcd.io/v2beta1
+    apiVersion: helm.toolkit.fluxcd.io/v1
     kind: HelmRelease
     name: podinfo
   environments:
     - name: dev
       targets:
         - namespace: dev-github
     - name: prod
       targets:
         - namespace: prod-github
   promotion:
```

### `ClusterBootstrapConfig`

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
