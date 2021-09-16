#!/bin/bash

echo "Delete all kustomizations"
kubectl delete -n $1 kustomizations.kustomize.toolkit.fluxcd.io --all
echo "Delete all gitrepositories"
kubectl delete -n $1 gitrepositories.source.toolkit.fluxcd.io --all
echo "Delete all helmrepositories"
kubectl delete -n $1 helmreleases.helm.toolkit.fluxcd.io --all
kubectl delete -n $1 helmcharts.source.toolkit.fluxcd.io --all
kubectl delete -n $1 helmrepositories.source.toolkit.fluxcd.io --all
echo "Delete any running applications"
kubectl delete apps -n $1 --all
echo "Delete all secrets (except weave-gitops-enterprise-credentials in $1)"
for s in $(kubectl get secrets -n $1 --field-selector metadata.name!=weave-gitops-enterprise-credentials| grep weave-gitops-|cut -d' ' -f1); do kubectl delete secrets $s -n $1; done
