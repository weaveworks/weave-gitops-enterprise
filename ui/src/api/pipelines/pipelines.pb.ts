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

export type ApprovePromotionRequest = {
  namespace?: string
  name?: string
  env?: string
  revision?: string
}

export type ApprovePromotionResponse = {
  pullRequestURL?: string
}

export type ListError = {
  namespace?: string
  message?: string
}

export type ListPullRequestsRequest = {
  pipelineName?: string
  pipelineNamespace?: string
}

export type ListPullRequestsResponse = {
  pullRequests?: {[key: string]: string}
}

export class Pipelines {
  static ListPipelines(req: ListPipelinesRequest, initReq?: fm.InitReq): Promise<ListPipelinesResponse> {
    return fm.fetchReq<ListPipelinesRequest, ListPipelinesResponse>(`/v1/pipelines?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetPipeline(req: GetPipelineRequest, initReq?: fm.InitReq): Promise<GetPipelineResponse> {
    return fm.fetchReq<GetPipelineRequest, GetPipelineResponse>(`/v1/namespaces/${req["namespace"]}/pipelines/${req["name"]}?${fm.renderURLSearchParams(req, ["namespace", "name"])}`, {...initReq, method: "GET"})
  }
  static ApprovePromotion(req: ApprovePromotionRequest, initReq?: fm.InitReq): Promise<ApprovePromotionResponse> {
    return fm.fetchReq<ApprovePromotionRequest, ApprovePromotionResponse>(`/v1/pipelines/approve/${req["name"]}`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ListPullRequests(req: ListPullRequestsRequest, initReq?: fm.InitReq): Promise<ListPullRequestsResponse> {
    return fm.fetchReq<ListPullRequestsRequest, ListPullRequestsResponse>(`/v1/pipelines/list-prs/${req["pipelineName"]}`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}