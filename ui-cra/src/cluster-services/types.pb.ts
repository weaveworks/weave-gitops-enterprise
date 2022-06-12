/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/
export type Pagination = {
  pageSize?: number
  pageToken?: string
}

export type ListError = {
  clusterName?: string
  namespace?: string
  message?: string
}

export type Canary = {
  namespace?: string
  name?: string
  clusterName?: string
  provider?: string
  targetReference?: CanaryTargetReference
  targetDeployment?: CanaryTargetDeployment
  status?: CanaryStatus
  deploymentStrategy?: string
}

export type CanaryTargetReference = {
  kind?: string
  name?: string
}

export type CanaryStatus = {
  phase?: string
  failedChecks?: number
  canaryWeight?: number
  iterations?: number
  lastTransitionTime?: string
  conditions?: CanaryCondition[]
}

export type CanaryCondition = {
  type?: string
  status?: string
  lastUpdateTime?: string
  lastTransitionTime?: string
  reason?: string
  message?: string
}

export type CanaryTargetDeployment = {
  uid?: string
  resourceVersion?: string
  fluxLabels?: FluxLabels
}

export type FluxLabels = {
  kustomizeNamespace?: string
  kustomizeName?: string
}

export type Automation = {
  kind?: string
  name?: string
  namespace?: string
}