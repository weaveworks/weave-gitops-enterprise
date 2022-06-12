/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "./fetch.pb"
import * as Types from "./types.pb"
export type GetVersionRequest = {
}

export type GetVersionResponse = {
  version?: string
}

export type ListCanariesRequest = {
  clusterName?: string
  pagination?: Types.Pagination
}

export type ListCanariesResponse = {
  canaries?: Types.Canary[]
  nextPageToken?: string
  errors?: Types.ListError[]
}

export type GetCanaryRequest = {
  name?: string
  namespace?: string
  clusterName?: string
}

export type GetCanaryResponse = {
  canary?: Types.Canary
  automation?: Types.Automation
}

export type IsFlaggerAvailableRequest = {
}

export type IsFlaggerAvailableResponse = {
  clusters?: {[key: string]: boolean}
}

export class ProgressiveDeliveryService {
  static GetVersion(req: GetVersionRequest, initReq?: fm.InitReq): Promise<GetVersionResponse> {
    return fm.fetchReq<GetVersionRequest, GetVersionResponse>(`/v1/version?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ListCanaries(req: ListCanariesRequest, initReq?: fm.InitReq): Promise<ListCanariesResponse> {
    return fm.fetchReq<ListCanariesRequest, ListCanariesResponse>(`/v1/canaries?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetCanary(req: GetCanaryRequest, initReq?: fm.InitReq): Promise<GetCanaryResponse> {
    return fm.fetchReq<GetCanaryRequest, GetCanaryResponse>(`/v1/canaries/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static IsFlaggerAvailable(req: IsFlaggerAvailableRequest, initReq?: fm.InitReq): Promise<IsFlaggerAvailableResponse> {
    return fm.fetchReq<IsFlaggerAvailableRequest, IsFlaggerAvailableResponse>(`/v1/crd/flagger?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
}