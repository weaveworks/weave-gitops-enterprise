# Cost estimation

## Introduction

The cost estimation feature allows you to get a sense of how much a CAPI template's rendered output will cost. It is a best-effort estimate, and is not guaranteed to be accurate. It is based on the AWS Pricing API and is only available for AWS.

## Availability of this feature

Cost estimation was requested by a specific customer (IQT) and is not available for general use for now. These instructions are for internal use only.

## Enabling

We make some configuration changes to the `values` in the Weave Gitops Enterprise `HelmRelease`.

```yaml
spec:
  values:
    config:
      costEstimation:
        # QueryString of filters to apply to cost estimation
        estimationFilter: "operatingSystem=Linux&tenancy=Dedicated&capacityStatus=UnusedCapacityReservation&operation=RunInstances"
        # the region used to fetch the pricing data, us-east-1 or ap-south-1 are publicly documented
        # as the only regions that support the pricing API
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
2. _(optional)_ over-ride the global defaults on a per-template basis with the `weave.works/cost-estimation-filter` annotation on the `Cluster` resource in the template.
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
