/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "./fetch.pb"
export type ListTemplatesRequest = {
}

export type ListTemplatesResponse = {
  templates?: Template[]
  total?: number
}

export type ListTemplateParamsRequest = {
  templateName?: string
}

export type ListTemplateParamsResponse = {
  parameters?: Parameter[]
}

export type RenderTemplateRequest = {
  templateName?: string
  values?: ParameterValues
}

export type RenderTemplateResponse = {
  renderedTemplate?: string
}

export type Template = {
  name?: string
  description?: string
  version?: string
  parameters?: Parameter[]
}

export type Parameter = {
  name?: string
  description?: string
}

export type ParameterValues = {
  values?: {[key: string]: string}
}

export class ClustersService {
  static ListTemplates(req: ListTemplatesRequest, initReq?: fm.InitReq): Promise<ListTemplatesResponse> {
    return fm.fetchReq<ListTemplatesRequest, ListTemplatesResponse>(`/v1/templates?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ListTemplateParams(req: ListTemplateParamsRequest, initReq?: fm.InitReq): Promise<ListTemplateParamsResponse> {
    return fm.fetchReq<ListTemplateParamsRequest, ListTemplateParamsResponse>(`/v1/templates/${req["templateName"]}/params?${fm.renderURLSearchParams(req, ["templateName"])}`, {...initReq, method: "GET"})
  }
  static RenderTemplate(req: RenderTemplateRequest, initReq?: fm.InitReq): Promise<RenderTemplateResponse> {
    return fm.fetchReq<RenderTemplateRequest, RenderTemplateResponse>(`/v1/templates/${req["templateName"]}/render`, {...initReq, method: "POST", body: JSON.stringify(req["values"])})
  }
}