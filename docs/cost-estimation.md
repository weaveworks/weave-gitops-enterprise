# Cost estimation

## Introduction

The cost estimation feature allows you to get a sense of how much a CAPI template's rendered output will cost. It is a best-effort estimate, and is not guaranteed to be accurate. It is based on the AWS Pricing API and is only available for AWS.

## Availability of this feature

Cost estimation was requested by a specific customer (IQT) and is not available for general use for now. These instructions are for internal use only.

## Enabling

We make some configuration changes to the `values` in the Weave GitOps Enterprise `HelmRelease`.

### AWS CSV Pricer

```yaml
spec:
  values:
    extraEnvVars:
      - name: WEAVE_GITOPS_FEATURE_COST_ESTIMATION
        value: "true"
      - name: COST_ESTIMATION_CSV_FILE
        value: "/var/aws-pricing/aws-pricing.csv"
    extraVolumeMounts:
      - name: aws-pricing
        mountPath: /var/aws-pricing
        readOnly: true
    extraVolumes:
      - name: aws-pricing
        configMap:
          name: aws-pricing
```

Given a sample `aws-pricing.csv` file:

```csv
currency,serviceCode,regionCode,instanceType,price
USD,AmazonEC2,us-east-1,t3.large,0.1
USD,AmazonEC2,us-east-1,t3.medium,0.08
```

Create a configmap.yaml with the following command:

```bash
kubectl create configmap aws-pricing \
  --namespace flux-system \
  --from-file=aws-pricing.csv \
  --dry-run=client -o yaml > configmap.yaml
```

Either save the configmap to git for flux to apply it, or apply it directly:

```bash
kubectl apply -f configmap.yaml
```

> **Warning**
>
> When updating the pricing data `ConfigMap` you must restart the cluster-service to refresh the new values and see the updated estimates in the UI.
>
> ```
> kubectl -n flux-system rollout restart deploy/weave-gitops-enterprise-mccp-cluster-service
> ```


### AWS Pricer

```yaml
spec:
  values:
    config:
      costEstimation:
        # QueryString of filters to apply to cost estimation
        estimationFilter: "operatingSystem=Linux&tenancy=Dedicated&capacityStatus=UnusedCapacityReservation&operation=RunInstances"
        # the region used to fetch the pricing data, us-east-1 or ap-south-1 are publicly documented
        # as the only regions that support the pricing API.
        # In the TopSecret region this is documented as only being available in the eastern region us-iso-east-1.
        apiRegion: us-east-1

    extraEnvVars:
      - name: WEAVE_GITOPS_FEATURE_COST_ESTIMATION
        value: "true"

    # If the cluster is not configured with IAM roles for service accounts, the following
    # line is required to provide the AWS credentials to the cost estimation service
    extraEnvVarsSecret: "aws-pricing-api-credentials"
```

AWS credentials can be provided explicitly in the `extraEnvVarsSecret` or implicitly by using IAM roles for service accounts. If provided explicitly, the secret can be created with the following command:

```bash
kubectl create secret generic aws-pricing-api-credentials \
  --namespace flux-system \
  --from-literal=AWS_ACCESS_KEY_ID='$MY_ACCESS_KEY' \
  --from-literal=AWS_SECRET_ACCESS_KEY='$MY_SECRET_KEY'
```

### Permissions

The cost estimation service requires the `AWSPriceListServiceFullAccess` policy attached to the user or role used to access the pricing API.

### Rollout

After updating the `HelmRelease`, commit, push, reconcile and restart:

- `git commit -am "Enable cost estimation"`
- `git push`
- `flux reconcile kustomization --with-source flux-system`
- `kubectl -n flux-system rollout restart deploy/weave-gitops-enterprise-mccp-cluster-service`

## Estimation

The estimate is only based on the machine pool of the CAPI cluster, we take into account:

- The number of control planes
- The number of worker nodes
- The deployment region
- Instance types

`instanceType` and `region` are automatically added to the estimation filters based on the values found in the template. Other filters should be added to get a more accurate estimate.

### Configuring the estimation filters

Providing additional filters based on your AWS environment will improve the accuracy of the estimate. For example in internal testing we have been using the following filters:

- `operatingSystem=Linux`
- `tenancy=Dedicated`
- `capacityStatus=UnusedCapacityReservation`
- `operation=RunInstances`

Estimation filters are configured both via:

1. _(required)_ Global defaults in the `HelmRelease`:
   ```yaml
   config:
     costEstimation:
       estimationFilter: "operatingSystem=Linux&tenancy=Dedicated&capacityStatus=UnusedCapacityReservation&operation=RunInstances"
   ```
2. _(optional)_ over-ride the global defaults on a per-template basis with the `weave.works/cost-estimation-filter` annotation on the `Cluster` resource in the template. Its important to put the annotation on the `Cluster` resource inside the template and not the `CAPITemplate` itself.

   ```yaml
   apiVersion: cluster.x-k8s.io/v1beta1
   kind: Cluster
   metadata:
     name: "test-cluster"
   annotations:
     "templates.weave.works/estimation-filters": "tenancy=Shared"
   ```

In both cases the filters should be expressed in the form of a query string.

## Hiding the cost estimation button on certain templates

You can disable the cost-estimation button from appearing on certain templates by adding the `templates.weave.works/cost-estimation-enabled: "false"` annotation to the `CAPITemplate`.

## Things that are out of scope

The following are out of scope for the cost estimation feature and may produce errors or inaccurate estimates if you try to estimate templates that use them:

- EKS clusters, managed and unmanaged
- Auto Scaling groups of spot EC2 instances
- CAPI ClusterClasses

Things that are ignored and are not included in the estimate:

- Other cluster resources like:
  - NAT gateways
  - Load balancers
  - Databases
  - Storage
  - etc
- MachinePool `minSize` and `maxSize` are ignored, the estimate is based on the number of `replicas`

# Appendix 1: Example CAPA template

```yaml
apiVersion: templates.weave.works/v1alpha2
kind: GitOpsTemplate
metadata:
  name: capa-official-cluster-template
  namespace: default
spec:
  params:
    - name: CLUSTER_NAME
      description: This is used for the cluster naming.
    - name: AWS_SSH_KEY_NAME
      description: AWS SSH key name
      default: default
    - name: KUBERNETES_VERSION
      description: Kubernetes version
      options:
        - v1.22.1
        - v1.23.2
      default: v1.23.2
    - name: CONTROL_PLANE_MACHINE_COUNT
      description: Number of control planes
      options:
        - "3"
        - "5"
      default: "3"
    - name: AWS_CONTROL_PLANE_MACHINE_TYPE
      description: Machine type
      options:
        - t3.medium
        - t3.large
      default: t3.medium
    - name: WORKER_MACHINE_COUNT
      description: Number of control planes
      options:
        - "3"
        - "4"
        - "5"
        - "6"
        - "7"
        - "8"
        - "9"
      default: "3"
    - name: AWS_NODE_MACHINE_TYPE
      description: Machine type
      options:
        - t3.nano
        - t3.micro
        - t3.small
        - t3.medium
        - t3.large
        - t3.xlarge
        - t3.2xlarge
      default: t3.large
    - name: AWS_REGION
      description: AWS Region
      options:
        - us-east-2
        - us-east-1
        - us-west-1
        - us-west-2
        - af-south-1
        - ap-east-1
        - ap-south-1
        - ap-northeast-3
        - ap-northeast-2
        - ap-southeast-1
        - ap-southeast-2
        - ap-northeast-1
        - ca-central-1
        - eu-central-1
        - eu-west-1
        - eu-west-2
        - eu-south-1
        - eu-west-3
        - eu-north-1
        - me-south-1
        - sa-east-1
      default: us-east-1
  resourcetemplates:
    - content:
        - apiVersion: cluster.x-k8s.io/v1beta1
          kind: Cluster
          metadata:
            name: "${CLUSTER_NAME}"
          spec:
            clusterNetwork:
              pods:
                cidrBlocks:
                  - 192.168.0.0/16
            infrastructureRef:
              apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
              kind: AWSCluster
              name: "${CLUSTER_NAME}"
            controlPlaneRef:
              kind: KubeadmControlPlane
              apiVersion: controlplane.cluster.x-k8s.io/v1beta1
              name: "${CLUSTER_NAME}-control-plane"
        - apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
          kind: AWSCluster
          metadata:
            name: "${CLUSTER_NAME}"
          spec:
            region: "${AWS_REGION}"
            sshKeyName: "${AWS_SSH_KEY_NAME}"
        - kind: KubeadmControlPlane
          apiVersion: controlplane.cluster.x-k8s.io/v1beta1
          metadata:
            name: "${CLUSTER_NAME}-control-plane"
          spec:
            replicas: "${CONTROL_PLANE_MACHINE_COUNT}"
            machineTemplate:
              infrastructureRef:
                kind: AWSMachineTemplate
                apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
                name: "${CLUSTER_NAME}-control-plane"
            kubeadmConfigSpec:
              initConfiguration:
                nodeRegistration:
                  name: "{{ ds.meta_data.local_hostname }}"
                  kubeletExtraArgs:
                    cloud-provider: aws
              clusterConfiguration:
                apiServer:
                  extraArgs:
                    cloud-provider: aws
                controllerManager:
                  extraArgs:
                    cloud-provider: aws
              joinConfiguration:
                nodeRegistration:
                  name: "{{ ds.meta_data.local_hostname }}"
                  kubeletExtraArgs:
                    cloud-provider: aws
            version: "${KUBERNETES_VERSION}"
        - kind: AWSMachineTemplate
          apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
          metadata:
            name: "${CLUSTER_NAME}-control-plane"
          spec:
            template:
              spec:
                instanceType: "${AWS_CONTROL_PLANE_MACHINE_TYPE}"
                iamInstanceProfile: control-plane.cluster-api-provider-aws.sigs.k8s.io
                sshKeyName: "${AWS_SSH_KEY_NAME}"
        - apiVersion: cluster.x-k8s.io/v1beta1
          kind: MachineDeployment
          metadata:
            name: "${CLUSTER_NAME}-md-0"
          spec:
            clusterName: "${CLUSTER_NAME}"
            replicas: "${WORKER_MACHINE_COUNT}"
            selector:
              matchLabels: null
            template:
              spec:
                clusterName: "${CLUSTER_NAME}"
                version: "${KUBERNETES_VERSION}"
                bootstrap:
                  configRef:
                    name: "${CLUSTER_NAME}-md-0"
                    apiVersion: bootstrap.cluster.x-k8s.io/v1beta1
                    kind: KubeadmConfigTemplate
                infrastructureRef:
                  name: "${CLUSTER_NAME}-md-0"
                  apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
                  kind: AWSMachineTemplate
        - apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
          kind: AWSMachineTemplate
          metadata:
            name: "${CLUSTER_NAME}-md-0"
          spec:
            template:
              spec:
                instanceType: "${AWS_NODE_MACHINE_TYPE}"
                iamInstanceProfile: nodes.cluster-api-provider-aws.sigs.k8s.io
                sshKeyName: "${AWS_SSH_KEY_NAME}"
        - apiVersion: bootstrap.cluster.x-k8s.io/v1beta1
          kind: KubeadmConfigTemplate
          metadata:
            name: "${CLUSTER_NAME}-md-0"
          spec:
            template:
              spec:
                joinConfiguration:
                  nodeRegistration:
                    name: "{{ ds.meta_data.local_hostname }}"
                    kubeletExtraArgs:
                      cloud-provider: aws
```

# Appendix: Available filters

This is a list of the available filters.

- instanceCapacityMetal
- volumeType
- maxIopsvolume
- instance
- classicnetworkingsupport
- instanceCapacity10xlarge
- fromRegionCode
- locationType
- toLocationType
- instanceFamily
- operatingSystem
- toRegionCode
- clockSpeed
- LeaseContractLength
- ecu
- networkPerformance
- instanceCapacity8xlarge
- group
- maxThroughputvolume
- gpuMemory
- ebsOptimized
- vpcnetworkingsupport
- maxVolumeSize
- gpu
- intelAvxAvailable
- processorFeatures
- instanceCapacity4xlarge
- servicecode
- groupDescription
- elasticGraphicsType
- volumeApiName
- processorArchitecture
- physicalCores
- fromLocation
- snapshotarchivefeetype
- marketoption
- availabilityzone
- productFamily
- fromLocationType
- enhancedNetworkingSupported
- intelTurboAvailable
- memory
- dedicatedEbsThroughput
- vcpu
- OfferingClass
- instanceCapacityLarge
- capacitystatus
- termType
- storage
- toLocation
- intelAvx2Available
- storageMedia
- regionCode
- physicalProcessor
- provisioned
- servicename
- PurchaseOption
- instancesku
- productType
- instanceCapacity18xlarge
- instanceType
- tenancy
- usagetype
- normalizationSizeFactor
- instanceCapacity16xlarge
- instanceCapacity2xlarge
- maxIopsBurstPerformance
- instanceCapacity12xlarge
- instanceCapacity32xlarge
- instanceCapacityXlarge
- licenseModel
- currentGeneration
- preInstalledSw
- transferType
- location
- instanceCapacity24xlarge
- instanceCapacity9xlarge
- instanceCapacityMedium
- operation
- resourceType
