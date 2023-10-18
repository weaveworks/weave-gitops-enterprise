/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"
import * as PreviewV1Types from "./types.pb"
export class PreviewService {
  static GetYAML(req: PreviewV1Types.GetYAMLRequest, initReq?: fm.InitReq): Promise<PreviewV1Types.GetYAMLResponse> {
    return fm.fetchReq<PreviewV1Types.GetYAMLRequest, PreviewV1Types.GetYAMLResponse>(`/v1/preview-yaml`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}