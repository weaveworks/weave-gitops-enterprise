/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "./fetch.pb"
import * as GoogleApiHttpbody from "./google/api/httpbody.pb"
import * as GoogleProtobufAny from "./google/protobuf/any.pb"
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

export type ListGitopsClustersRequest = {
  label?: string
  pageSize?: string
  pageToken?: string
}

export type ListGitopsClustersResponse = {
  gitopsClusters?: GitopsCluster[]
  total?: number
  nxtPageToken?: string
}

export type GetPolicyRequest = {
  policyName?: string
}

export type ListPoliciesRequest = {
}

export type GetPolicyResponse = {
  policy?: Policy
}

export type ListPoliciesResponse = {
  policies?: Policy[]
  total?: number
}

export type ListPolicyValidationsRequest = {
  clusterId?: string
}

export type ListPolicyValidationsResponse = {
  violations?: PolicyValidation[]
  total?: number
}

export type GetPolicyValidationRequest = {
  violationId?: string
}

export type GetPolicyValidationResponse = {
  violation?: PolicyValidation
}

export type PolicyValidation = {
  id?: string
  message?: string
  clusterId?: string
  category?: string
  severity?: string
  createdAt?: string
  entity?: string
  namespace?: string
  violatingEntity?: string
  description?: string
  howToSolve?: string
  name?: string
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

export type Condition = {
  type?: string
  status?: string
  reason?: string
  message?: string
  timestamp?: string
}

export type GitopsCluster = {
  name?: string
  namespace?: string
  annotations?: {[key: string]: string}
  labels?: {[key: string]: string}
  conditions?: Condition[]
  capiClusterRef?: GitopsClusterRef
  secretRef?: GitopsClusterRef
  capiCluster?: CapiCluster
}

export type CapiCluster = {
  name?: string
  namespace?: string
  annotations?: {[key: string]: string}
  labels?: {[key: string]: string}
  status?: CapiClusterStatus
}

export type CapiClusterStatus = {
  phase?: string
  infrastructureReady?: boolean
  controlPlaneInitialized?: boolean
  controlPlaneReady?: boolean
  conditions?: Condition[]
  observedGeneration?: string
}

export type GitopsClusterRef = {
  name?: string
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

export type PolicyParamRepeatedString = {
  value?: string[]
}

export type PolicyParam = {
  name?: string
  type?: string
  value?: GoogleProtobufAny.Any
  required?: boolean
}

export type PolicyTargetLabel = {
  values?: {[key: string]: string}
}

export type PolicyTargets = {
  kinds?: string[]
  labels?: PolicyTargetLabel[]
  namespaces?: string[]
}

export type Policy = {
  name?: string
  id?: string
  code?: string
  description?: string
  howToSolve?: string
  category?: string
  tags?: string[]
  severity?: string
  controls?: string[]
  gitCommit?: string
  parameters?: PolicyParam[]
  targets?: PolicyTargets
  createdAt?: string
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
  static ListGitopsClusters(req: ListGitopsClustersRequest, initReq?: fm.InitReq): Promise<ListGitopsClustersResponse> {
    return fm.fetchReq<ListGitopsClustersRequest, ListGitopsClustersResponse>(`/v1/clusters?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
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
  static ListPolicies(req: ListPoliciesRequest, initReq?: fm.InitReq): Promise<ListPoliciesResponse> {
    return fm.fetchReq<ListPoliciesRequest, ListPoliciesResponse>(`/v1/policies?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetPolicy(req: GetPolicyRequest, initReq?: fm.InitReq): Promise<GetPolicyResponse> {
    return fm.fetchReq<GetPolicyRequest, GetPolicyResponse>(`/v1/policies/${req["policyName"]}?${fm.renderURLSearchParams(req, ["policyName"])}`, {...initReq, method: "GET"})
  }
  static ListPolicyValidations(req: ListPolicyValidationsRequest, initReq?: fm.InitReq): Promise<ListPolicyValidationsResponse> {
    return fm.fetchReq<ListPolicyValidationsRequest, ListPolicyValidationsResponse>(`/v1/policyviolations`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetPolicyValidation(req: GetPolicyValidationRequest, initReq?: fm.InitReq): Promise<GetPolicyValidationResponse> {
    return fm.fetchReq<GetPolicyValidationRequest, GetPolicyValidationResponse>(`/v1/policyviolations/${req["violationId"]}?${fm.renderURLSearchParams(req, ["violationId"])}`, {...initReq, method: "GET"})
  }
}