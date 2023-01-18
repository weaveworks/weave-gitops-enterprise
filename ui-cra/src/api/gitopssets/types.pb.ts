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
}

export type ListError = {
  namespace?: string
  message?: string
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