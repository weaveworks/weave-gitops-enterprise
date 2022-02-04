# MCCP

Helm chart to install the MCCP (Multi-cluster control plane) component.

## TL;DR

This chart uses images from private Docker repositories so you will need to supply valid docker credentials in order to use them. You can find instructions on how to create a secret based on existing Docker credentials [here](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

In addition to these images, the chart also installs [NATS](https://github.com/bitnami/charts/tree/master/bitnami/nats) and [NGINX Ingress Controller](https://github.com/bitnami/charts/tree/master/bitnami/nginx-ingress-controller). Since NATS is exposed directly as a NodePort service in the Kubernetes cluster (in the default configuration), it is often convenient to specify which port to use as well as the address that NATS is accessible from.

```bash
$ helm repo add wkp s3://weaveworks-wkp/charts-v3
$ helm install my-release wkp/mccp \
    --set "imagePullSecrets[0].name=<secret-containing-docker-config>" \
    --set "nats.client.service.nodePort=<exposed-port-for-nats>" \
    --set "agentTemplate.natsURL=<nats-address>:<exposed-port-for-nats>"
```

> Note: When using `sqlite` as the backing store, you need to designate which worker node will be used to host the database file otherwise the pods will stay in `Pending` state. Run `kubectl label nodes <node-name> wkp-database-volume-node=true` to add the label to the designated node. Once the label has been added to a node the pods should transition into the `Running` state.

## Prerequisites

- Kubernetes 1.16+
- Helm 3.5.4

## Installing the Chart

To install the chart with the release name `my-release`:

```bash
$ helm install my-release wkp/mccp
```

The command deploys MCCP on the Kubernetes cluster in the default configuration.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```bash
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.
