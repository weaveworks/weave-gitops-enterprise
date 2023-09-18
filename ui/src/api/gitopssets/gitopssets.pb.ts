/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"
import * as GitopssetsV1Types from "./types.pb"
export type ListGitOpsSetsRequest = {
  namespace?: string
}

export type ListGitOpsSetsResponse = {
  gitopssets?: GitopssetsV1Types.GitOpsSet[]
  errors?: GitopssetsV1Types.GitOpsSetListError[]
}

export type ToggleSuspendGitOpsSetRequest = {
  clusterName?: string
  name?: string
  namespace?: string
  suspend?: boolean
  clusterNamespace?: string
}

export type ToggleSuspendGitOpsSetResponse = {
}

export type GetReconciledObjectsRequest = {
  name?: string
  namespace?: string
  automationKind?: string
  kinds?: GitopssetsV1Types.GroupVersionKind[]
  clusterName?: string
  clusterNamespace?: string
}

export type GetReconciledObjectsResponse = {
  objects?: GitopssetsV1Types.Object[]
}

export type SyncGitOpsSetRequest = {
  clusterName?: string
  name?: string
  namespace?: string
}

export type SyncGitOpsSetResponse = {
  success?: boolean
}

export type GetGitOpsSetRequest = {
  name?: string
  namespace?: string
  clusterNamespace?: string
  clusterName?: string
}

export type GetGitOpsSetResponse = {
  gitopsSet?: GitopssetsV1Types.GitOpsSet
}

export class GitOpsSets {
  static ListGitOpsSets(req: ListGitOpsSetsRequest, initReq?: fm.InitReq): Promise<ListGitOpsSetsResponse> {
    return fm.fetchReq<ListGitOpsSetsRequest, ListGitOpsSetsResponse>(`/v1/gitopssets?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetGitOpsSet(req: GetGitOpsSetRequest, initReq?: fm.InitReq): Promise<GetGitOpsSetResponse> {
    return fm.fetchReq<GetGitOpsSetRequest, GetGitOpsSetResponse>(`/v1/namespaces/${req["clusterNamespace"]}/clusters/${req["clusterName"]}/namespaces/${req["namespace"]}/gitopssets/${req["name"]}?${fm.renderURLSearchParams(req, ["clusterNamespace", "clusterName", "namespace", "name"])}`, {...initReq, method: "GET"})
  }
  static ToggleSuspendGitOpsSet(req: ToggleSuspendGitOpsSetRequest, initReq?: fm.InitReq): Promise<ToggleSuspendGitOpsSetResponse> {
    return fm.fetchReq<ToggleSuspendGitOpsSetRequest, ToggleSuspendGitOpsSetResponse>(`/v1/namespaces/${req["clusterNamespace"]}/clusters/${req["clusterName"]}/namespaces/${req["namespace"]}/gitopssets/suspend`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetReconciledObjects(req: GetReconciledObjectsRequest, initReq?: fm.InitReq): Promise<GetReconciledObjectsResponse> {
    return fm.fetchReq<GetReconciledObjectsRequest, GetReconciledObjectsResponse>(`/v1/namespaces/${req["clusterNamespace"]}/clusters/${req["clusterName"]}/namespaces/${req["namespace"]}/gitops-reconciled-objects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static SyncGitOpsSet(req: SyncGitOpsSetRequest, initReq?: fm.InitReq): Promise<SyncGitOpsSetResponse> {
    return fm.fetchReq<SyncGitOpsSetRequest, SyncGitOpsSetResponse>(`/v1/gitopssets/sync`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}