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
  errors?: GitopssetsV1Types.ListError[]
}

export class GitOpsSets {
  static ListGitOpsSets(req: ListGitOpsSetsRequest, initReq?: fm.InitReq): Promise<ListGitOpsSetsResponse> {
    return fm.fetchReq<ListGitOpsSetsRequest, ListGitOpsSetsResponse>(`/v1/gitopssets?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
}