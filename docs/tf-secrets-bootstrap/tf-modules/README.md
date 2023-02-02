# Bootstrapping secrets to leaf cluster using terraform

## Requirements

- Working management cluster and a leaf cluster.
- TF-Controller installed on management cluster [see here](../apps/tf-controller/release.yaml).
- Service account on AWS to authenticate the TF-Controller.
- AdminRoleBinding for TF-Controller. [see here](../apps/tf-controller/rolebinding.yaml)
- Cluster template with Terraform object points to tf-modules [see here](../clusters/waleed-terraform/capi-templates/aws-eks.yaml). This can be changed according to your setup
- Leaf cluster [terraform modules](https://github.com/weaveworks/clusters-config/tree/cluster-waleed-terraform/eksctl-clusters/leaf-terraform) (sync-secrets, flux, external-secrets) managed by this module [here](main.tf)
- Github token for flux to be able to push files

## Steps

1- Make your changes in the template according to your setup then add it to your cluster and make sure it's installed

```bash
➜ k get CAPITemplate
NAMESPACE   NAME      AGE
default     aws-eks   6h40m

```

2- Create a service account with a proper access [see here](https://docs.gitops.weave.works/docs/terraform/aws-eks/) to authenticate tf-controller

3- Install TF-Controller on the management cluster by making a kustomization to point to its location

- Kustomization file:

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1beta1
kind: Kustomization
metadata:
  name: tf-controller
  namespace: flux-system
spec:
  interval: 30s
  sourceRef:
    kind: GitRepository
    name: flux-system
  path: ./eksctl-clusters/apps/tf-controller
  prune: true
  validation: client
```

- Helm Repo & Helm Release and make sure to replace with the proper role

```yaml
---
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: HelmRepository
metadata:
  name: tf-controller
  namespace: flux-system
spec:
  interval: 1h0s
  url: https://weaveworks.github.io/tf-controller/
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: tf-controller
  namespace: flux-system
spec:
  chart:
    spec:
      chart: tf-controller
      sourceRef:
        kind: HelmRepository
        name: tf-controller
      version: ">=0.9.3"
  interval: 1h0s
  releaseName: tf-controller
  targetNamespace: flux-system
  install:
    crds: Create
  upgrade:
    crds: CreateReplace
  values:
    replicaCount: 1
    concurrency: 24
    resources:
      limits:
        cpu: 1000m
        memory: 2Gi
      requests:
        cpu: 400m
        memory: 64Mi
    caCertValidityDuration: 24h
    certRotationCheckFrequency: 30m
    image:
      tag: v0.13.1
    runner:
      image:
        repository: ghcr.io/rparmer/tf-runner
        tag: ubuntu-1669733504
      serviceAccount:
        annotations:
          eks.amazonaws.com/role-arn: "arn:aws:iam::894516026745:role/leaf-tf-controller" # TODO: replace with your role
    awsPackage:
      install: true
      tag: v4.38.0-v1alpha11
```

- TF-Controller Role binding file

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tf-runner-admin
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
subjects:
  - kind: ServiceAccount
    name: tf-runner
    namespace: flux-system
```

3- Configure the leaf-terraform modules with your required values (modify github branch, tokens, ...)

4- Create a new cluster using the template we created before

5- Wait for tf-controller job to finish

**Note (To be fixed)**

- you may have the following error

```bash
│ The "for_each" map includes keys derived from resource attributes that cannot be determined until apply, and so Terraform cannot determine the full set of keys that will
│ identify the instances of this resource.
│
│ When working with unknown values in for_each, it's better to define the map keys statically in your configuration and place apply-time results only in the map values.
│
│ Alternatively, you could use the -target planning option to first apply only the resources that the for_each value depends on, and then apply a second time to fully converge.
╵
```

- This is because the flux-system have to be applied before flux installation

Temp fix using

```bash
➜  flux git:(add-flux) ✗ tf apply -var "github_token=$GITHUB_TOKEN" --target=kubernetes_namespace.flux_system                                                          <aws:sts>
kubernetes_namespace.flux_system: Refreshing state... [id=wge2205-leaf-flux-system]

No changes. Your infrastructure matches the configuration.
```

Then the job should be fixed
