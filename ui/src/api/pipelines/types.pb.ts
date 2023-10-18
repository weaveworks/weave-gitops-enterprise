/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/
export type ClusterRef = {
  kind?: string
  name?: string
  namespace?: string
}

export type Target = {
  namespace?: string
  clusterRef?: ClusterRef
}

export type Environment = {
  name?: string
  targets?: Target[]
  promotion?: Promotion
}

export type ObjectRef = {
  kind?: string
  name?: string
  namespace?: string
}

export type AppRef = {
  apiVersion?: string
  kind?: string
  name?: string
}

export type Condition = {
  type?: string
  status?: string
  reason?: string
  message?: string
  timestamp?: string
}

export type WorkloadStatus = {
  kind?: string
  name?: string
  version?: string
  lastAppliedRevision?: string
  conditions?: Condition[]
  suspended?: boolean
}

export type PipelineTargetStatus = {
  clusterRef?: ClusterRef
  namespace?: string
  workloads?: WorkloadStatus[]
}

export type WaitingStatus = {
  revision?: string
}

export type PipelineStatusEnvironmentStatus = {
  waitingStatus?: WaitingStatus
  targetsStatuses?: PipelineTargetStatus[]
}

export type PipelineStatus = {
  environments?: {[key: string]: PipelineStatusEnvironmentStatus}
}

export type Pipeline = {
  name?: string
  namespace?: string
  appRef?: AppRef
  environments?: Environment[]
  targets?: Target[]
  status?: PipelineStatus
  yaml?: string
  type?: string
  promotion?: Promotion
}

export type PullRequestList = {
  pullRequests?: PullRequest[]
}

export type PullRequest = {
  title?: string
  url?: string
}

export type Promotion = {
  manual?: boolean
  strategy?: Strategy
}

export type Strategy = {
  pullRequest?: PullRequestPromotion
  notification?: Notification
  secretRef?: LocalObjectReference
}

export type PullRequestPromotion = {
  type?: string
  url?: string
  branch?: string
}

export type Notification = {
}

export type LocalObjectReference = {
  name?: string
}