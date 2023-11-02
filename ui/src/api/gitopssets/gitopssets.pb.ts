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
}

export type ToggleSuspendGitOpsSetResponse = {
}

export type GetReconciledObjectsRequest = {
  name?: string
  namespace?: string
  automationKind?: string
  kinds?: GitopssetsV1Types.GroupVersionKind[]
  clusterName?: string
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
    return fm.fetchReq<GetGitOpsSetRequest, GetGitOpsSetResponse>(`/v1/namespaces/${req["namespace"]}/gitopssets/${req["name"]}?${fm.renderURLSearchParams(req, ["namespace", "name"])}`, {...initReq, method: "GET"})
  }
  static ToggleSuspendGitOpsSet(req: ToggleSuspendGitOpsSetRequest, initReq?: fm.InitReq): Promise<ToggleSuspendGitOpsSetResponse> {
    return fm.fetchReq<ToggleSuspendGitOpsSetRequest, ToggleSuspendGitOpsSetResponse>(`/v1/namespaces/${req["namespace"]}/gitopssets/${req["name"]}/suspend`, {...initReq, method: "PATCH", body: JSON.stringify(req, fm.replacer)})
  }
  static GetReconciledObjects(req: GetReconciledObjectsRequest, initReq?: fm.InitReq): Promise<GetReconciledObjectsResponse> {
    return fm.fetchReq<GetReconciledObjectsRequest, GetReconciledObjectsResponse>(`/v1/namespaces/${req["namespace"]}/gitopssets/${req["name"]}/reconciled-objects`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static SyncGitOpsSet(req: SyncGitOpsSetRequest, initReq?: fm.InitReq): Promise<SyncGitOpsSetResponse> {
    return fm.fetchReq<SyncGitOpsSetRequest, SyncGitOpsSetResponse>(`/v1/namespaces/${req["namespace"]}/gitopssets/${req["name"]}/sync`, {...initReq, method: "PATCH", body: JSON.stringify(req, fm.replacer)})
  }
}