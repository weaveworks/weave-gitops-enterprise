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
}

export type GetPipelineRequest = {
  name?: string
  namespace?: string
}

export type GetPipelineResponse = {
  pipeline?: PipelinesV1Types.Pipeline
}

export type GetPipelineStatusForObjectRequest = {
  object?: PipelinesV1Types.ObjectRef
}

export type GetPipelineStatusForObjectResponse = {
  pipelineStatus?: PipelinesV1Types.PipelineStatus
}

export class Pipelines {
  static ListPipelines(req: ListPipelinesRequest, initReq?: fm.InitReq): Promise<ListPipelinesResponse> {
    return fm.fetchReq<ListPipelinesRequest, ListPipelinesResponse>(`/v1/pipelines?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetPipeline(req: GetPipelineRequest, initReq?: fm.InitReq): Promise<GetPipelineResponse> {
    return fm.fetchReq<GetPipelineRequest, GetPipelineResponse>(`/v1/pipelines/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static GetPipelineStatusForObject(req: GetPipelineStatusForObjectRequest, initReq?: fm.InitReq): Promise<GetPipelineStatusForObjectResponse> {
    return fm.fetchReq<GetPipelineStatusForObjectRequest, GetPipelineStatusForObjectResponse>(`/v1/pipelines/status?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
}