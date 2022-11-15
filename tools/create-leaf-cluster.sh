#!/bin/bash
set -e

if [ $# -eq 0 ]
then
    echo "Please, provide the cluster name (e.g. leaf-cluster-01)"
    exit
fi

if ! command -v vcluster &> /dev/null
then
    echo "vcluster could not be found. You can install it by following https://www.vcluster.com/docs/getting-started/setup"
    exit
fi

echo "Creating cluster..."
vcluster create --connect=false -n vcluster-$1 $1

echo "Waiting cluster config..."
TRY=12
until [[ $i -gt $TRY  ]] || kubectl get secret -n vcluster-$1 vc-$1 &>/dev/null
do
    sleep 10
    i=$((i+1))
done

echo "Creating GitopsCluster secret..."
kubectl get secret -n vcluster-$1 vc-$1 --template={{.data.config}} \
 | base64 --decode \
 | sed "s/localhost:8443/$1.vcluster-$1/g" \
 | kubectl create secret -n vcluster-$1 generic $1-config --from-file=value=/dev/stdin

echo "Creating GitopsCluster resource..."
cat <<EOF | kubectl apply -f -
apiVersion: gitops.weave.works/v1alpha1
kind: GitopsCluster
metadata:
  name: $1
  namespace: vcluster-$1
  # Signals that this cluster should be bootstrapped.
  labels:
    weave.works/capi: bootstrap
spec:
  secretRef:
    name: $1-config
EOF

function clean_up() {
    echo "Disconnecting from cluster"
    vcluster disconnect
}
trap clean_up EXIT

echo "Waiting for cluster to be ready..."
i=0
until [[ $i -gt $TRY  ]] || vcluster list -n vcluster-$1 | grep -q Running
do
    sleep 10
    i=$((i+1))
done

echo "Connecting to cluster"
vcluster connect -n vcluster-$1 $1

echo "Installing demo workload"
flux install

kubectl apply -f test/utils/data/user-roles.yaml
kubectl apply -f test/utils/data/admin-role-bindings.yaml

flux create source helm podinfo --namespace=default --url=https://stefanprodan.github.io/podinfo --interval=10m
flux create helmrelease podinfo --namespace=default --source=HelmRepository/podinfo --release-name=podinfo --chart=podinfo
