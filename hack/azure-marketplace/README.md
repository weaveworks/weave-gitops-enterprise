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

- cli tools, all these can be installed via `brew install` on mac:
  - `az`
  - `docker`
  - `python`
  - `helm`
  - `crane`
- Access to the weaveworks engineering Azure account
  - TODO: add instructions for how to get access

## Push the images to ACR

Login to the Azure

```bash
az login
az acr login -n weaveworksmarketplacepublic.azurecr.io
```

```bash
helm repo add weave-gitops-enterprise-charts https://charts.dev.wkp.weave.works/releases/charts-v3
helm repo update

# this branch, maybe will merge to main one day
git checkout aws-marketplace
# where you have the weave-gitops-enterprise repo
cd ~/weave-gitops-enterprise
cd ./hack/azure-marketplace

./publish_images_azure.py --version 0.25.0 --local-helm-chart weave-gitops-enterprise-charts --dry-run
```

Check that it looks sort of sensible then run it for real

```bash
./publish_images_azure.py --version 0.25.0 --local-helm-chart weave-gitops-enterprise-charts
```

Head to the [weaveworksmarketplacepublic.azurecr.io registry on the Azure portal](https://portal.azure.com/#@Weave365.onmicrosoft.com/resource/subscriptions/ace37984-3d07-4051-9002-d5a52c0ae14b/resourceGroups/team-pesto-use1/providers/Microsoft.ContainerRegistry/registries/weaveworksmarketplacepublic/repository) and check that the images are there

## Convert some weave-gitops-enterprise helm chart to an azure compatible chart

Grab a chart from some weaveworks helm repo

```bash
# add if you haven't already
helm repo add weave-gitops-enterprise-charts https://charts.dev.wkp.weave.works/releases/charts-v3
helm repo update

# then pull the chart
helm pull weave-gitops-enterprise-charts/mccp --version 0.25.0
```

Extract the chart and inspect it

```bash
# remove the existing chart
rm -rf ./mccp/*

# -xtractzeefiles back out to recreate the chart
tar -xzf mccp-0.25.0.tgz

# Have a look at the base chart to see the diff between the "normal" chart, and the azure chart we just removed
git diff
```

Convert it to an azure compatible chart

```bash
# up, if you haven't already
# convert the chart to an azure compatible chart
python3 to_azure_chart.py mccp

# See whats changed
git diff
```

## Test the helm chart

> **Warning**
> FIXME: This doesn't work right now. `kind_load_images.py` is broken, see here -
> https://github.com/kubernetes-sigs/kind/issues/2394

Lets test with `kind`!

The standard prep:

1. `kind create cluster`
2. `flux install`
3. `kubectl apply -f ~/weave-gitops-enterprise/test/utils/data/entitlement/entitlement-secret.yaml`
4. `kubectl create secret generic cluster-user-auth -n flux-system --from-literal=username=wego-admin --from-literal=password='$2a$10$8zn1EXGdMcBPuE8a3BsxHeJzHS6b3s1YAXzWgdbZg4z8CDeG4wjJ6'` -- **gitops** is the password here
5. `clusterctl init --infrastructure vcluster`

Pull all the images from `global.azure.images` section from ACR, then push them into a kind cluster.

```
python3 kind_load_images.py ./mccp
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
# (if you're not already in the hack/azure-marketplace dir..)
cd ./hack/azure-marketplace

docker run -it -v /var/run/docker.sock:/var/run/docker.sock -v `pwd`:/data --entrypoint "/bin/bash" mcr.microsoft.com/container-package-app:latest

export REGISTRY_NAME=weaveworksmarketplacepublic.azurecr.io
az login
az acr login -n $REGISTRY_NAME
cd /data/azure

cpa verify

# Push it up to the registry
cpa buildbundle
```

Once this is done you should be able to see the new CNAB in the [weave.works.weave-gitops-enterprise repository on ACR](https://portal.azure.com/#view/Microsoft_Azure_ContainerRegistries/RepositoryBlade/id/%2Fsubscriptions%2Face37984-3d07-4051-9002-d5a52c0ae14b%2FresourceGroups%2Fteam-pesto-use1%2Fproviders%2FMicrosoft.ContainerRegistry%2Fregistries%2Fweaveworksmarketplacepublic/repository/weave.works.weave-gitops-enterprise)

Shut down the container, commit the new chart to git and push it up

```bash
git add .
git commit -m "Add published chart for 0.24.3"
git push
```

## Update the marketplace offering

> **Warning**
> The marketplace UI is a bit funky and the "Add CNAB bundle" button mentioned here often stays disabled, reloading the browser a few times usually fixes it. If the previous CNAB bundle was not actually published (because it had a CVE detected in it etc), then you might need to delete the existing CNAB bundle before you can add the new one. Maybe.

1. Head to https://partner.microsoft.com/en-us/dashboard/commercial-marketplace/offers/c4ad183b-9eb3-4c72-b237-969f1ebcf6e7/plans/a30c11a0-ec49-408b-9818-fcd2675bb104/technicalconfiguration
1. If the ids have changed by the time you're following these instructions then maybe head to https://partner.microsoft.com/en-us/dashboard/commercial-marketplace and then to the bit about the "technical" component of the offering.
1. Add the new CNAB bundle (public marketplace > weave.works.weave-gitops-enterprise > 0.24.3)
1. Save draft
1. Review and publish, I've been leaving the "notes for reviewer" section empty so far.

## User-configuration

Azure does not allow the user to configure values.yaml directly, instead a UI must be defined for each variable the user should be allows to configure.

**This work has not been done** so instead the User can only make minimal configuration changes to the cluster-service environment by an optional configmap.

# User-guide

see [./user-guide.md](./user-guide.md)

## Modifying the createUIDefinition.json

If you're modifying the createUIDefinition.json you can test it out in the "Sandbox" before publishing: https://portal.azure.com/#view/Microsoft_Azure_CreateUIDef/SandboxBlade

Click through the steps and see if the changes you've made look alright. At the end of the wizard there is an option to see all the generated values that will be passed to the mainTemplate.json too.

## Oh no! An image got flagged as vulnerable during marketplace publishing! CVEs..

The "quick and easy" way to do this is

1. Address the CVE in the container image, for example, some go lib in WGE needs to be bumped we:
   1. Add the `replace` directive to the go.mod file
2. Get a new docker image
   1. Either build it locally or push a branch to github and let it build the images
   2. Go find where it built and pushed that image to on dockerhub/ghcr by looking at the github actions logs
3. Get the new image up to ACR (after logging in), choosing some sensible new tag, if `2.5.0` has the CVE, maybe push as `2.5.1-rc.1`.
   1. Copy from weaveworks CI: `crane cp weaveworks/weave-gitops-enterprise-clusters-service:fix-azure-cve-circl-4ffa57ce weaveworksmarketplacepublic.azurecr.io/weave-gitops-enterprise-clusters-service:v0.28.1-rc.1`
   2. Push a locally rebuild image with `docker tag docker.io/library/policy-agent:452.f950f4b weaveworksmarketplacepublic.azurecr.io/policy-agent:v2.6.0-rc.1`, `docker push weaveworksmarketplacepublic.azurecr.io/policy-agent:v2.6.0-rc.1`)
4. Find the new sha via the ACR UI
5. Update ./hack/azure-marketplace/mccp/values.yaml with the new sha
6. Follow the **Build and push it up to the ACR registry** above, bumping the patch version again, the Azure version will be our of sync with the WGE version but that's ok.
