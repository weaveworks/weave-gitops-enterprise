# Publishing onto the Azure Marketplace

This document describes how to publish the Weave GitOps Enterprise product onto the Azure Marketplace.

## Understand

We need to provide a CNAB (porter (?)) bundle, a “mega docker image” including the EE helm chart and all dependent docker images.

#### Images

- All images included in this CNAB need to be read from Azure container registry (ACR) as we build it
- All images will be referenced by their SHA256, not their tag

#### Helm chart

- Azure requires a certain format of helm chart, with a values.yaml that specifies all images used in deployments under a `global.azure` key

### EE artifact refresher

- Stores docker images for deployments on docker hub and ghcr
- Stores helm charts in s3 buckets
- `values.yaml` specifies image tags in many different custom ways

## Tasks

Given the above, how can we take an EE helm chart and get the equivilant CNAB that can be attached to an azure marketplace container offering:

- Get an existing, released EE helm chart
- Determine what images are included in it
- Copy the images from docker hub and ghcr to Azure Container Registry (ACR)
- Rewrite the helm chart so every deployment.yaml image field reads image information from `global.azure` in values.yaml
- Test this new helm chart
- Build and push a CNAB

## Prerequisites

- cli tools:
  - `az`
  - `docker`
  - `python`
  - `helm`
- Access to the weaveworks engineering Azure account
  - TODO: add instructions for how to get access

## Convert some weave-gitops-enterprise helm chart to an azure compatible chart

Grab a chart from some weaveworks helm repo

```bash
helm repo add wkpv3 https://s3.us-east-1.amazonaws.com/weaveworks-wkp/charts-v3/
helm repo update
helm pull wkpv3/mccp --version 0.23.0-63-gf2634d2
```

Extract the chart and inspect it

```bash
# -xtractzeefiles
tar -xzf mccp-0.23.0-63-gf2634d2.tgz
cd mccp
git init
git add .
git commit -m "initial commit"
```

TODO: determine what images are in there, and copy to ACR

Login to the Azure

```bash
az login
az acr login -n weaveworksmarketplacepublic.azurecr.io
```

Convert it to an azure compatible chart

```bash
# up, if you haven't already
cd ..
# convert the chart to an azure compatible chart
python3 to_azure_chart.py mccp
cd mccp

# See whats changed
git diff
```

## Test the helm chart

Lets test with `kind`!

The standard prep:

1. `kind create cluster`
2. `flux install`
3. `kubectl apply -f ~/weave-gitops-enterprise/test/utils/data/entitlement/entitlement-secret.yaml`
4. `kubectl create secret generic cluster-user-auth -n flux-system --from-literal=username=wego-admin --from-literal=password='$2a$10$8zn1EXGdMcBPuE8a3BsxHeJzHS6b3s1YAXzWgdbZg4z8CDeG4wjJ6'` -- **gitops** is the password here
5. `clusterctl init --infrastructure vcluster`

Pull all the images from `global.azure.images` section from ACR, then push them into a kind cluster.

```
az login
az acr login -n weaveworksmarketplacepublic.azurecr.io
python3 load_kind_images.py ./mccp
```

Install the chart

```
helm install --namespace flux-system weave-gitops-enterprise ./mccp
```

See if its working:

```
kubectl get pods -A
kubectl port-forward --namespace flux-system svc/clusters-service 8000:8000
```

Open https://localhost:8000 (https!)

## Build and push it up to the ACR registry

Bump the version in ./azure/manifest.yaml

_If you don't do this it will refuse to overwrite the existing tag when running `cpa buildbundle`. Overwriting existing tags with `--force` doesn't seem to work that well when going through marketplace re-validation etc, its hard to know if the changes have come through. So just increment it, it might be out of sync with the official release numbers but ho hum._

```bash
# up, if you haven't already
cd ..
vim ./azure/manifest.yaml

# edit the version field to something like 0.24.3
# NOTE: it **cannot** be an RC (e.g. 0.24.0-rc.1) or it won't appear in the marketplace
```

So now we'll actually push it up, we're going to roughly follow the instructions here:

- https://learn.microsoft.com/en-us/partner-center/marketplace/azure-container-technical-assets-kubernetes?tabs=linux#manually-run-the-packaging-tool

```bash
docker run -it -v /var/run/docker.sock:/var/run/docker.sock -v `pwd`:/data --entrypoint "/bin/bash" mcr.microsoft.com/container-package-app:latest

export REGISTRY_NAME=weaveworksmarketplacepublic.azurecr.io
az login
az acr login -n $REGISTRY_NAME
cd /data/azure

cpa verify

# Push it up to the registry
cpa buildbundle
```

## Update the marketplace offering

1. Head to https://partner.microsoft.com/en-us/dashboard/commercial-marketplace/offers/c4ad183b-9eb3-4c72-b237-969f1ebcf6e7/plans/4c72a3fe-e39e-44a6-9720-bae241dce038/technicalconfiguration
2. If the ids have changed by the time you're following these instructions then maybe head to https://partner.microsoft.com/en-us/dashboard/commercial-marketplace and then to the bit about the "technical" component of the offering.
3. Remove the existing CNAB bundle
4. Add the new CNAB bundle (public marketplace > weave.works.weave-gitops-enterprise > 0.24.3)
5. Save draft
6. Review and publish, I've been leaving the "notes for reviewer" section empty so far.

## User-configuration

Azure does not allow the user to configure values.yaml directly, instead a UI must be defined for each variable the user should be allows to configure.

**This work has not been done** so instead the User can only make minimal configuration changes to the cluster-service environment by an optional configmap.

# User-guide

see [./user-guide.md](./user-guide.md)
