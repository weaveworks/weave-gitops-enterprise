/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/
export type SourceRef = {
  apiVersion?: string
  kind?: string
  name?: string
  namespace?: string
}

export type Interval = {
  hours?: string
  minutes?: string
  seconds?: string
}

export type ResourceRef = {
  name?: string
  type?: string
  identifier?: string
}

export type NamespacedObjectReference = {
  name?: string
  namespace?: string
}

export type ObjectRef = {
  clusterName?: string
  name?: string
  namespace?: string
  kind?: string
}

export type TerraformObject = {
  name?: string
  namespace?: string
  clusterName?: string
  sourceRef?: SourceRef
  appliedRevision?: string
  path?: string
  interval?: Interval
  lastUpdatedAt?: string
  driftDetectionResult?: boolean
  inventory?: ResourceRef[]
  conditions?: Condition[]
  labels?: {[key: string]: string}
  annotations?: {[key: string]: string}
  dependsOn?: NamespacedObjectReference[]
  suspended?: boolean
}

export type Pagination = {
  pageSize?: number
  pageToken?: string
}

export type TerraformListError = {
  clusterName?: string
  namespace?: string
  message?: string
}

export type Condition = {
  type?: string
  status?: string
  reason?: string
  message?: string
  timestamp?: string
}