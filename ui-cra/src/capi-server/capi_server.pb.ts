/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "./fetch.pb"
import * as GoogleApiHttpbody from "./google/api/httpbody.pb"
export type ListTemplatesRequest = {
  provider?: string
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

export type ListTemplateProfilesRequest = {
  templateName?: string
}

export type ListTemplateProfilesResponse = {
  profiles?: TemplateProfile[]
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

export type ListWeaveClustersRequest = {
  label?: string
}

export type ListWeaveClustersResponse = {
  clusters?: Cluster[]
  total?: number
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
  values?: ProfileValues[]
  repositoryApiUrl?: string
}

export type CreatePullRequestResponse = {
  webUrl?: string
}

export type DeleteClustersPullRequestRequest = {
  repositoryUrl?: string
  headBranch?: string
  baseBranch?: string
  title?: string
  description?: string
  clusterNames?: string[]
  commitMessage?: string
  credentials?: Credential
  repositoryApiUrl?: string
}

export type DeleteClustersPullRequestResponse = {
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

export type Cluster = {
  name?: string
  type?: string
  status?: string
  label?: string
  objects?: ClusterObject[]
  error?: string
}

export type ClusterObject = {
  kind?: string
  apiVersion?: string
  name?: string
  displayName?: string
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
  provider?: string
  parameters?: Parameter[]
  objects?: TemplateObject[]
  error?: string
  annotations?: {[key: string]: string}
}

export type Parameter = {
  name?: string
  description?: string
  required?: boolean
  options?: string[]
}

export type TemplateProfile = {
  name?: string
  version?: string
}

export type TemplateObject = {
  kind?: string
  apiVersion?: string
  parameters?: string[]
  name?: string
  displayName?: string
}

export type GetEnterpriseVersionRequest = {
}

export type GetEnterpriseVersionResponse = {
  version?: string
}

export type Maintainer = {
  name?: string
  email?: string
  url?: string
}

export type HelmRepository = {
  name?: string
  namespace?: string
}

export type Profile = {
  name?: string
  home?: string
  sources?: string[]
  description?: string
  keywords?: string[]
  maintainers?: Maintainer[]
  icon?: string
  annotations?: {[key: string]: string}
  kubeVersion?: string
  helmRepository?: HelmRepository
  availableVersions?: string[]
}

export type ProfileValues = {
  name?: string
  version?: string
  values?: string
  layer?: string
}

export type GetConfigRequest = {
}

export type GetConfigResponse = {
  repositoryURL?: string
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
  static ListTemplateProfiles(req: ListTemplateProfilesRequest, initReq?: fm.InitReq): Promise<ListTemplateProfilesResponse> {
    return fm.fetchReq<ListTemplateProfilesRequest, ListTemplateProfilesResponse>(`/v1/templates/${req["templateName"]}/profiles?${fm.renderURLSearchParams(req, ["templateName"])}`, {...initReq, method: "GET"})
  }
  static RenderTemplate(req: RenderTemplateRequest, initReq?: fm.InitReq): Promise<RenderTemplateResponse> {
    return fm.fetchReq<RenderTemplateRequest, RenderTemplateResponse>(`/v1/templates/${req["templateName"]}/render`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListWeaveClusters(req: ListWeaveClustersRequest, initReq?: fm.InitReq): Promise<ListWeaveClustersResponse> {
    return fm.fetchReq<ListWeaveClustersRequest, ListWeaveClustersResponse>(`/v1/clusters?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static CreatePullRequest(req: CreatePullRequestRequest, initReq?: fm.InitReq): Promise<CreatePullRequestResponse> {
    return fm.fetchReq<CreatePullRequestRequest, CreatePullRequestResponse>(`/v1/clusters`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static DeleteClustersPullRequest(req: DeleteClustersPullRequestRequest, initReq?: fm.InitReq): Promise<DeleteClustersPullRequestResponse> {
    return fm.fetchReq<DeleteClustersPullRequestRequest, DeleteClustersPullRequestResponse>(`/v1/clusters`, {...initReq, method: "DELETE", body: JSON.stringify(req)})
  }
  static ListCredentials(req: ListCredentialsRequest, initReq?: fm.InitReq): Promise<ListCredentialsResponse> {
    return fm.fetchReq<ListCredentialsRequest, ListCredentialsResponse>(`/v1/credentials?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetKubeconfig(req: GetKubeconfigRequest, initReq?: fm.InitReq): Promise<GoogleApiHttpbody.HttpBody> {
    return fm.fetchReq<GetKubeconfigRequest, GoogleApiHttpbody.HttpBody>(`/v1/clusters/${req["clusterName"]}/kubeconfig?${fm.renderURLSearchParams(req, ["clusterName"])}`, {...initReq, method: "GET"})
  }
  static GetEnterpriseVersion(req: GetEnterpriseVersionRequest, initReq?: fm.InitReq): Promise<GetEnterpriseVersionResponse> {
    return fm.fetchReq<GetEnterpriseVersionRequest, GetEnterpriseVersionResponse>(`/v1/enterprise/version?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetConfig(req: GetConfigRequest, initReq?: fm.InitReq): Promise<GetConfigResponse> {
    return fm.fetchReq<GetConfigRequest, GetConfigResponse>(`/v1/config?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
}