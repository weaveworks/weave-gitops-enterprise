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

export type GetGitOpsSetRequest = {
  name?: string
  namespace?: string
  kind?: string
}

export type GetGitOpsSetResponse = {
  gitopsSet?: GitopssetsV1Types.GitOpsSet
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

export type GetChildObjectsRequest = {
  groupVersionKind?: GitopssetsV1Types.GroupVersionKind
  namespace?: string
  parentUid?: string
  clusterName?: string
}

export type GetChildObjectsResponse = {
  objects?: GitopssetsV1Types.Object[]
}

export class GitOpsSets {
  static ListGitOpsSets(req: ListGitOpsSetsRequest, initReq?: fm.InitReq): Promise<ListGitOpsSetsResponse> {
    return fm.fetchReq<ListGitOpsSetsRequest, ListGitOpsSetsResponse>(`/v1/gitopssets?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetGitOpsSet(req: GetGitOpsSetRequest, initReq?: fm.InitReq): Promise<GetGitOpsSetResponse> {
    return fm.fetchReq<GetGitOpsSetRequest, GetGitOpsSetResponse>(`/v1/gitopsset/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static ToggleSuspendGitOpsSet(req: ToggleSuspendGitOpsSetRequest, initReq?: fm.InitReq): Promise<ToggleSuspendGitOpsSetResponse> {
    return fm.fetchReq<ToggleSuspendGitOpsSetRequest, ToggleSuspendGitOpsSetResponse>(`/v1/gitopssets/suspend`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetReconciledObjects(req: GetReconciledObjectsRequest, initReq?: fm.InitReq): Promise<GetReconciledObjectsResponse> {
    return fm.fetchReq<GetReconciledObjectsRequest, GetReconciledObjectsResponse>(`/v1/reconciled_objects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetChildObjects(req: GetChildObjectsRequest, initReq?: fm.InitReq): Promise<GetChildObjectsResponse> {
    return fm.fetchReq<GetChildObjectsRequest, GetChildObjectsResponse>(`/v1/child_objects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}