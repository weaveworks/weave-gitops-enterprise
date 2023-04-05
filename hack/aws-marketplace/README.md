# Publishing onto the AWS Marketplace

This document describes how to publish the Weave GitOps Enterprise product onto the AWS Marketplace.

## Prerequisites

- cli tools:
  - `aws`
  - `gsts`
  - `docker`
  - `python`
  - `yq`
  - `helm`
- MarketplaceAdmin IAM role

## Steps

Login to the AWS Marketplace Admin account and switch to the MarketplaceAdmin role.

```bash
export AWS_PROFILE=sts
export AWS_ROLE_MARKETPLACE="arn:aws:iam::677537422032:role/MarketplaceAdmin"
export GOOGLE_IDP_ID=C0203uytv
export GOOGLE_SP_ID=656726301855
gsts --aws-role-arn "$AWS_ROLE_MARKETPLACE" --sp-id "$GOOGLE_SP_ID" --idp-id "$GOOGLE_IDP_ID" --username simon@weave.works
```

Test it out, (nothing is pushed)

```bash
./publish_images.py --version 0.20.0 \
    --aws-image-name weave-gitops-enterprise-development \
    --local-helm-chart weave-gitops-enterprise-charts \
    --dry-run
```

Inspect the generated helm archive etc

```
tar -xf weave-gitops-enterprise-development-0.20.1-rc.1.tgz
grep -R ecr weave-gitops-enterprise-development/values.yaml
```

Then actually publish images and oci helm chart:

```bash
./publish_images.py --version 0.20.0 \
    --aws-image-name weave-gitops-enterprise-development \
    --local-helm-chart weave-gitops-enterprise-charts
```

## Installing

### Helm cli

While the ECR repo is private you'll have to manually load in the images into kind

```bash
./kind_load_images.py weave-gitops-enterprise-development-0.20.1-rc.1.tgz
```

Then helm install

```bash
helm install --set global.capiEnabled=false --namespace flux-system mccp ./weave-gitops-enterprise-development-0.20.1-rc.1.tgz
```

### Flux CRs

Or with a helm release (should be able to skip the auth if the ECR repo has been published and is public)

```bash
flux create secret oci ecr-auth \
    -n flux-system \
    --url 709825985650.dkr.ecr.us-east-1.amazonaws.com/weaveworks
    --username AWS \
    --password=$(aws ecr get-login-password --region us-east-1)
```

and load the yaml

```yaml
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: weave-gitops-enterprise-production
  namespace: flux-system
spec:
  interval: 10m
  type: oci
  url: oci://709825985650.dkr.ecr.us-east-1.amazonaws.com/weaveworks
  secretRef:
    name: ecr-auth
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: weave-gitops-enterprise-production
  namespace: flux-system
spec:
  interval: 10m
  targetNamespace: flux-system
  # release-name is important as the default one is too long
  releaseName: mccp
  chart:
    spec:
      chart: weave-gitops-enterprise-production
      version: 0.20.1-rc.1
      sourceRef:
        kind: HelmRepository
        name: weave-gitops-enterprise-production
  install:
    crds: CreateReplace
  upgrade:
    crds: CreateReplace
  interval: 50m
  values:
    global:
      capiEnabled: false
```
