/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"
import * as PipelinesV1Types from "./types.pb"
export type ListPipelinesRequest = {
  namespace?: string
}

export type ListPipelinesResponse = {
  pipelines?: PipelinesV1Types.Pipeline[]
  errors?: ListError[]
}

export type GetPipelineRequest = {
  name?: string
  namespace?: string
}

export type GetPipelineResponse = {
  pipeline?: PipelinesV1Types.Pipeline
  errors?: string[]
}

export type ListError = {
  namespace?: string
  message?: string
}

export class Pipelines {
  static ListPipelines(req: ListPipelinesRequest, initReq?: fm.InitReq): Promise<ListPipelinesResponse> {
    return fm.fetchReq<ListPipelinesRequest, ListPipelinesResponse>(`/v1/pipelines?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetPipeline(req: GetPipelineRequest, initReq?: fm.InitReq): Promise<GetPipelineResponse> {
    return fm.fetchReq<GetPipelineRequest, GetPipelineResponse>(`/v1/pipelines/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
}