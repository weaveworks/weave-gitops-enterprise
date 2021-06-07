/* eslint-disable */
// @ts-nocheck
/*
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb"
export type RenderTemplateRequest = {
  templateName?: string
  params?: {[key: string]: string}
}

export type RenderTemplateResponse = {
  renderedTemplate?: string
}

export type ListTemplateParamsRequest = {
  templateName?: string
}

export type ListTemplateParamsResponse = {
  params?: Param[]
}

export type Param = {
  name?: string;
  description?: string;
};

export type Template = {
  name?: string;
  description?: string;
  version?: string;
  params?: Param[];
};

export type ListTemplatesRequest = {
}

export type ListTemplatesResponse = {
  templates?: Template[];
  total?: number;
};

export class ClustersService {
  static ListTemplates(
    req: ListTemplatesRequest,
    initReq?: fm.InitReq,
  ): Promise<ListTemplatesResponse> {
    return fm.fetchReq<ListTemplatesRequest, ListTemplatesResponse>(
      `/v1/templates?${fm.renderURLSearchParams(req, [])}`,
      { ...initReq, method: 'GET' },
    );
  }
  static ListTemplateParams(req: ListTemplateParamsRequest, initReq?: fm.InitReq): Promise<ListTemplateParamsResponse> {
    return fm.fetchReq<ListTemplateParamsRequest, ListTemplateParamsResponse>(`/v1/templates/${req["templateName"]}/params?${fm.renderURLSearchParams(req, ["templateName"])}`, {...initReq, method: "GET"})
  }
  static RenderTemplate(req: RenderTemplateRequest, initReq?: fm.InitReq): Promise<RenderTemplateResponse> {
    return fm.fetchReq<RenderTemplateRequest, RenderTemplateResponse>(`/v1/templates/${req["templateName"]}/render`, {...initReq, method: "POST"})
  }
}
