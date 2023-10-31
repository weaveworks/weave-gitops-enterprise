/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"
import * as PreviewV1Types from "./types.pb"
export class PreviewService {
  static GetYAML(req: PreviewV1Types.GetYAMLRequest, initReq?: fm.InitReq): Promise<PreviewV1Types.GetYAMLResponse> {
    return fm.fetchReq<PreviewV1Types.GetYAMLRequest, PreviewV1Types.GetYAMLResponse>(`/v1/preview/yaml`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static CreatePullRequest(req: PreviewV1Types.CreatePullRequestRequest, initReq?: fm.InitReq): Promise<PreviewV1Types.CreatePullRequestResponse> {
    return fm.fetchReq<PreviewV1Types.CreatePullRequestRequest, PreviewV1Types.CreatePullRequestResponse>(`/v1/preview/pull-requests`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}