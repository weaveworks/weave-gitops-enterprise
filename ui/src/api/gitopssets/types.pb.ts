/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/
export type GitOpsSet = {
  name?: string
  namespace?: string
  inventory?: ResourceRef[]
  conditions?: Condition[]
  clusterName?: string
  type?: string
  labels?: {[key: string]: string}
  annotations?: {[key: string]: string}
  suspended?: boolean
  observedGeneration?: string
  yaml?: string
  objectRef?: ObjectRef
}

export type GitOpsSetListError = {
  namespace?: string
  message?: string
  clusterName?: string
}

export type ResourceRef = {
  id?: string
  version?: string
}

export type Condition = {
  type?: string
  status?: string
  reason?: string
  message?: string
  timestamp?: string
}

export type GitOpsSetGenerator = {
  list?: string
  gitRepository?: string
}

export type GroupVersionKind = {
  group?: string
  kind?: string
  version?: string
}

export type Object = {
  payload?: string
  clusterName?: string
  tenant?: string
  uid?: string
  inventory?: GroupVersionKind[]
}

export type ObjectRef = {
  kind?: string
  name?: string
  namespace?: string
  clusterName?: string
}