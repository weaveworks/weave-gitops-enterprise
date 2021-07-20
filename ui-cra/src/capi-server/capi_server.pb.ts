/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "./fetch.pb"
import * as GoogleApiHttpbody from "./google/api/httpbody.pb"
export type ListTemplatesRequest = {
}

export type ListTemplatesResponse = {
  templates?: Template[]
  total?: number
}

export type GetTemplateRequest = {
  templateName?: string
}

export type GetTemplateResponse = {
  template?: Template
}

export type ListTemplateParamsRequest = {
  templateName?: string
}

export type ListTemplateParamsResponse = {
  parameters?: Parameter[]
  objects?: TemplateObject[]
}

export type RenderTemplateRequest = {
  templateName?: string
  values?: {[key: string]: string}
  credentials?: Credential
}

export type RenderTemplateResponse = {
  renderedTemplate?: string
}

export type CreatePullRequestRequest = {
  repositoryUrl?: string
  headBranch?: string
  baseBranch?: string
  title?: string
  description?: string
  templateName?: string
  parameterValues?: {[key: string]: string}
  commitMessage?: string
  credentials?: Credential
}

export type CreatePullRequestResponse = {
  webUrl?: string
}

export type ListCredentialsRequest = {
}

export type ListCredentialsResponse = {
  credentials?: Credential[]
  total?: number
}

export type GetKubeconfigRequest = {
  clusterName?: string
}

export type GetKubeconfigResponse = {
  kubeconfig?: string
}

export type Credential = {
  group?: string
  version?: string
  kind?: string
  name?: string
  namespace?: string
}

export type Template = {
  name?: string
  description?: string
  version?: string
  parameters?: Parameter[]
  body?: string
  objects?: TemplateObject[]
  error?: string
}

export type Parameter = {
  name?: string
  description?: string
  required?: boolean
  options?: string[]
}

export type TemplateObject = {
  kind?: string
  apiVersion?: string
  parameters?: string[]
}

export class ClustersService {
  static ListTemplates(req: ListTemplatesRequest, initReq?: fm.InitReq): Promise<ListTemplatesResponse> {
    return fm.fetchReq<ListTemplatesRequest, ListTemplatesResponse>(`/v1/templates?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetTemplate(req: GetTemplateRequest, initReq?: fm.InitReq): Promise<GetTemplateResponse> {
    return fm.fetchReq<GetTemplateRequest, GetTemplateResponse>(`/v1/templates/${req["templateName"]}?${fm.renderURLSearchParams(req, ["templateName"])}`, {...initReq, method: "GET"})
  }
  static ListTemplateParams(req: ListTemplateParamsRequest, initReq?: fm.InitReq): Promise<ListTemplateParamsResponse> {
    return fm.fetchReq<ListTemplateParamsRequest, ListTemplateParamsResponse>(`/v1/templates/${req["templateName"]}/params?${fm.renderURLSearchParams(req, ["templateName"])}`, {...initReq, method: "GET"})
  }
  static RenderTemplate(req: RenderTemplateRequest, initReq?: fm.InitReq): Promise<RenderTemplateResponse> {
    return fm.fetchReq<RenderTemplateRequest, RenderTemplateResponse>(`/v1/templates/${req["templateName"]}/render`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static CreatePullRequest(req: CreatePullRequestRequest, initReq?: fm.InitReq): Promise<CreatePullRequestResponse> {
    return fm.fetchReq<CreatePullRequestRequest, CreatePullRequestResponse>(`/v1/pulls`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListCredentials(req: ListCredentialsRequest, initReq?: fm.InitReq): Promise<ListCredentialsResponse> {
    return fm.fetchReq<ListCredentialsRequest, ListCredentialsResponse>(`/v1/credentials?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetKubeconfig(req: GetKubeconfigRequest, initReq?: fm.InitReq): Promise<GoogleApiHttpbody.HttpBody> {
    return fm.fetchReq<GetKubeconfigRequest, GoogleApiHttpbody.HttpBody>(`/v1/clusters/${req["clusterName"]}/kubeconfig?${fm.renderURLSearchParams(req, ["clusterName"])}`, {...initReq, method: "GET"})
  }
}