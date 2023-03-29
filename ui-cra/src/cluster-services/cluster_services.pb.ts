/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "./fetch.pb"
import * as GoogleApiHttpbody from "./google/api/httpbody.pb"
import * as GoogleProtobufAny from "./google/protobuf/any.pb"
import * as GoogleProtobufStruct from "./google/protobuf/struct.pb"
export type ListTemplatesRequest = {
  provider?: string
  templateKind?: string
}

export type Pagination = {
  pageSize?: number
  pageToken?: string
}

export type ListError = {
  clusterName?: string
  namespace?: string
  message?: string
}

export type ListTemplatesResponse = {
  templates?: Template[]
  total?: number
  errors?: ListError[]
}

export type GetTemplateRequest = {
  templateName?: string
  templateKind?: string
  templateNamespace?: string
}

export type GetTemplateResponse = {
  template?: Template
}

export type ListTemplateParamsRequest = {
  templateName?: string
  templateKind?: string
  templateNamespace?: string
}

export type ListTemplateParamsResponse = {
  parameters?: Parameter[]
  objects?: TemplateObject[]
}

export type ListTemplateProfilesRequest = {
  templateName?: string
  templateKind?: string
  templateNamespace?: string
}

export type ListTemplateProfilesResponse = {
  profiles?: TemplateProfile[]
  objects?: TemplateObject[]
}

export type RenderTemplateRequest = {
  templateName?: string
  values?: {[key: string]: string}
  credentials?: Credential
  templateKind?: string
  clusterNamespace?: string
  profiles?: ProfileValues[]
  kustomizations?: Kustomization[]
  templateNamespace?: string
  externalSecrets?: ExternalSecret[]
}

export type CommitFile = {
  path?: string
  content?: string
}

export type CostEstimateRange = {
  low?: number
  high?: number
}

export type CostEstimate = {
  currency?: string
  range?: CostEstimateRange
  message?: string
}

export type RenderTemplateResponse = {
  renderedTemplate?: CommitFile[]
  profileFiles?: CommitFile[]
  kustomizationFiles?: CommitFile[]
  costEstimate?: CostEstimate
  externalSecretsFiles?: CommitFile[]
  policyConfigFiles?: CommitFile[]
  sopsSecretFiles?: CommitFile[]
}

export type RenderAutomationRequest = {
  clusterAutomations?: ClusterAutomation[]
}

export type RenderAutomationResponse = {
  kustomizationFiles?: CommitFile[]
  helmReleaseFiles?: CommitFile[]
  externalSecretsFiles?: CommitFile[]
  policyConfigFiles?: CommitFile[]
  sopsSecertFiles?: CommitFile[]
}

export type ListGitopsClustersRequest = {
  label?: string
  pageSize?: string
  pageToken?: string
  refType?: string
}

export type ListGitopsClustersResponse = {
  gitopsClusters?: GitopsCluster[]
  total?: number
  nextPageToken?: string
  errors?: ListError[]
}

export type GetPolicyRequest = {
  policyName?: string
  clusterName?: string
}

export type ListPoliciesRequest = {
  clusterName?: string
  pagination?: Pagination
}

export type GetPolicyResponse = {
  policy?: Policy
  clusterName?: string
}

export type ListPoliciesResponse = {
  policies?: Policy[]
  total?: number
  nextPageToken?: string
  errors?: ListError[]
}

export type ListPolicyValidationsRequest = {
  clusterName?: string
  pagination?: Pagination
  application?: string
  namespace?: string
}

export type ListPolicyValidationsResponse = {
  violations?: PolicyValidation[]
  total?: number
  nextPageToken?: string
  errors?: ListError[]
}

export type GetPolicyValidationRequest = {
  violationId?: string
  clusterName?: string
}

export type GetPolicyValidationResponse = {
  violation?: PolicyValidation
}

export type PolicyValidationOccurrence = {
  message?: string
}

export type PolicyValidationParam = {
  name?: string
  type?: string
  value?: GoogleProtobufAny.Any
  required?: boolean
  configRef?: string
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
  clusterName?: string
  occurrences?: PolicyValidationOccurrence[]
  policyId?: string
  parameters?: PolicyValidationParam[]
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
  clusterNamespace?: string
  kustomizations?: Kustomization[]
  templateNamespace?: string
  templateKind?: string
  previousValues?: PreviousValues
  externalSecrets?: ExternalSecret[]
  policyConfigs?: PolicyConfigObject[]
  sopsSecrets?: SopsSecret[]
}

export type PreviousValues = {
  parameterValues?: {[key: string]: string}
  credentials?: Credential
  values?: ProfileValues[]
  kustomizations?: Kustomization[]
  externalSecrets?: ExternalSecret[]
  policyConfigs?: PolicyConfigObject[]
  sopsSecrets?: SopsSecret[]
}

export type CreatePullRequestResponse = {
  webUrl?: string
}

export type CreateTfControllerPullRequestRequest = {
  repositoryUrl?: string
  headBranch?: string
  baseBranch?: string
  title?: string
  description?: string
  templateName?: string
  parameterValues?: {[key: string]: string}
  commitMessage?: string
  repositoryApiUrl?: string
  templateNamespace?: string
}

export type CreateTfControllerPullRequestResponse = {
  webUrl?: string
}

export type ClusterNamespacedName = {
  namespace?: string
  name?: string
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
  clusterNamespacedNames?: ClusterNamespacedName[]
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
  clusterNamespace?: string
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
  controlPlane?: boolean
  type?: string
}

export type CapiCluster = {
  name?: string
  namespace?: string
  annotations?: {[key: string]: string}
  labels?: {[key: string]: string}
  status?: CapiClusterStatus
  infrastructureRef?: CapiClusterInfrastructureRef
}

export type CapiClusterStatus = {
  phase?: string
  infrastructureReady?: boolean
  controlPlaneInitialized?: boolean
  controlPlaneReady?: boolean
  conditions?: Condition[]
  observedGeneration?: string
}

export type CapiClusterInfrastructureRef = {
  apiVersion?: string
  kind?: string
  name?: string
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
  templateKind?: string
  templateType?: string
  labels?: {[key: string]: string}
  namespace?: string
  profiles?: TemplateProfile[]
}

export type Parameter = {
  name?: string
  description?: string
  required?: boolean
  options?: string[]
  default?: string
  editable?: boolean
}

export type TemplateProfile = {
  name?: string
  version?: string
  editable?: boolean
  values?: string
  namespace?: string
  required?: boolean
  profileTemplate?: string
  layer?: string
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

export type CreateAutomationsPullRequestRequest = {
  repositoryUrl?: string
  headBranch?: string
  baseBranch?: string
  title?: string
  description?: string
  commitMessage?: string
  repositoryApiUrl?: string
  clusterAutomations?: ClusterAutomation[]
}

export type ClusterAutomation = {
  cluster?: ClusterNamespacedName
  isControlPlane?: boolean
  kustomization?: Kustomization
  helmRelease?: HelmRelease
  filePath?: string
  externalSecret?: ExternalSecret
  policyConfig?: PolicyConfigObject
  sopsSecret?: SopsSecret
}

export type ExternalSecret = {
  metadata?: Metadata
  spec?: ExternalSecretSpec
}

export type ExternalSecretSpec = {
  refreshInterval?: string
  secretStoreRef?: externalSecretStoreRef
  target?: externalSecretTarget
  data?: externalSecretData
}

export type externalSecretStoreRef = {
  name?: string
  kind?: string
}

export type externalSecretTarget = {
  name?: string
}

export type externalSecretData = {
  secretKey?: string
  remoteRef?: externalSecretRemoteRef
}

export type externalSecretRemoteRef = {
  key?: string
  property?: string
}

export type Kustomization = {
  metadata?: Metadata
  spec?: KustomizationSpec
}

export type KustomizationSpec = {
  path?: string
  sourceRef?: SourceRef
  targetNamespace?: string
  createNamespace?: boolean
  decryption?: Decryption
}

export type Decryption = {
  provider?: string
  secretRef?: SecretRef
}

export type SecretRef = {
  name?: string
}

export type HelmRelease = {
  metadata?: Metadata
  spec?: HelmReleaseSpec
}

export type HelmReleaseSpec = {
  chart?: Chart
  values?: string
}

export type Chart = {
  spec?: ChartSpec
}

export type ChartSpec = {
  chart?: string
  sourceRef?: SourceRef
  version?: string
}

export type Metadata = {
  name?: string
  namespace?: string
  annotations?: {[key: string]: string}
}

export type SourceRef = {
  name?: string
  namespace?: string
}

export type CreateAutomationsPullRequestResponse = {
  webUrl?: string
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
  namespace?: string
}

export type GetConfigRequest = {
}

export type GetConfigResponse = {
  repositoryURL?: string
  managementClusterName?: string
  uiConfig?: string
  gitHostTypes?: {[key: string]: string}
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

export type PolicyStandard = {
  id?: string
  controls?: string[]
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
  standards?: PolicyStandard[]
  gitCommit?: string
  parameters?: PolicyParam[]
  targets?: PolicyTargets
  createdAt?: string
  clusterName?: string
  tenant?: string
  modes?: string[]
}

export type ObjectRef = {
  kind?: string
  name?: string
  namespace?: string
}

export type Event = {
  type?: string
  reason?: string
  message?: string
  timestamp?: string
  component?: string
  host?: string
  name?: string
}

export type ListEventsRequest = {
  involvedObject?: ObjectRef
  clusterName?: string
}

export type ListEventsResponse = {
  events?: Event[]
}

export type RepositoryRef = {
  cluster?: ClusterNamespacedName
  name?: string
  namespace?: string
  kind?: string
}

export type ListChartsForRepositoryRequest = {
  repository?: RepositoryRef
  kind?: string
}

export type RepositoryChart = {
  name?: string
  versions?: string[]
  layer?: string
}

export type ListChartsForRepositoryResponse = {
  charts?: RepositoryChart[]
}

export type GetValuesForChartRequest = {
  repository?: RepositoryRef
  name?: string
  version?: string
}

export type GetValuesForChartResponse = {
  jobId?: string
}

export type GetChartsJobRequest = {
  jobId?: string
}

export type GetChartsJobResponse = {
  values?: string
  error?: string
}

export type Workspace = {
  name?: string
  clusterName?: string
  namespaces?: string[]
}

export type ListWorkspacesRequest = {
  pagination?: Pagination
}

export type ListWorkspacesResponse = {
  workspaces?: Workspace[]
  total?: number
  nextPageToken?: string
  errors?: ListError[]
}

export type WorkspaceRoleRule = {
  groups?: string[]
  resources?: string[]
  verbs?: string[]
}

export type WorkspaceRole = {
  name?: string
  namespace?: string
  rules?: WorkspaceRoleRule[]
  manifest?: string
  timestamp?: string
}

export type WorkspaceRoleBindingRoleRef = {
  apiGroup?: string
  kind?: string
  name?: string
}

export type WorkspaceRoleBindingSubject = {
  apiGroup?: string
  kind?: string
  name?: string
  namespace?: string
}

export type WorkspaceRoleBinding = {
  name?: string
  namespace?: string
  manifest?: string
  timestamp?: string
  role?: WorkspaceRoleBindingRoleRef
  subjects?: WorkspaceRoleBindingSubject[]
}

export type WorkspaceServiceAccount = {
  name?: string
  namespace?: string
  timestamp?: string
  manifest?: string
}

export type WorkspacePolicy = {
  id?: string
  name?: string
  category?: string
  severity?: string
  timestamp?: string
}

export type GetWorkspaceRequest = {
  clusterName?: string
  workspaceName?: string
}

export type GetWorkspaceResponse = {
  name?: string
  clusterName?: string
  namespaces?: string[]
}

export type GetWorkspaceRolesResponse = {
  name?: string
  clusterName?: string
  objects?: WorkspaceRole[]
}

export type GetWorkspaceRoleBindingsResponse = {
  name?: string
  clusterName?: string
  objects?: WorkspaceRoleBinding[]
}

export type GetWorkspaceServiceAccountsResponse = {
  name?: string
  clusterName?: string
  objects?: WorkspaceServiceAccount[]
}

export type GetWorkspacePoliciesResponse = {
  name?: string
  clusterName?: string
  objects?: WorkspacePolicy[]
}

export type ExternalSecretItem = {
  secretName?: string
  externalSecretName?: string
  namespace?: string
  clusterName?: string
  secretStore?: string
  status?: string
  timestamp?: string
}

export type ListExternalSecretsRequest = {
}

export type ListExternalSecretsResponse = {
  secrets?: ExternalSecretItem[]
  total?: number
  errors?: ListError[]
}

export type GetExternalSecretRequest = {
  clusterName?: string
  namespace?: string
  externalSecretName?: string
}

export type GetExternalSecretResponse = {
  secretName?: string
  externalSecretName?: string
  clusterName?: string
  namespace?: string
  secretStore?: string
  secretStoreType?: string
  secretPath?: string
  property?: string
  version?: string
  status?: string
  timestamp?: string
}

export type ExternalSecretStore = {
  kind?: string
  name?: string
  namespace?: string
  type?: string
}

export type ListExternalSecretStoresRequest = {
  clusterName?: string
}

export type ListExternalSecretStoresResponse = {
  stores?: ExternalSecretStore[]
  total?: number
}

export type SyncExternalSecretsRequest = {
  clusterName?: string
  namespace?: string
  externalSecretName?: string
}

export type SyncExternalSecretsResponse = {
}

export type PolicyConfigListItem = {
  name?: string
  clusterName?: string
  totalPolicies?: number
  match?: string
  status?: string
  age?: string
}

export type ListPolicyConfigsRequest = {
}

export type ListPolicyConfigsResponse = {
  policyConfigs?: PolicyConfigListItem[]
  errors?: ListError[]
  total?: number
}

export type GetPolicyConfigRequest = {
  clusterName?: string
  name?: string
}

export type GetPolicyConfigResponse = {
  name?: string
  clusterName?: string
  age?: string
  status?: string
  matchType?: string
  match?: PolicyConfigMatch
  policies?: PolicyConfigPolicy[]
  totalPolicies?: number
}

export type PolicyConfigApplicationMatch = {
  name?: string
  kind?: string
  namespace?: string
}

export type PolicyConfigResourceMatch = {
  name?: string
  kind?: string
  namespace?: string
}

export type PolicyConfigMatch = {
  namespaces?: string[]
  workspaces?: string[]
  apps?: PolicyConfigApplicationMatch[]
  resources?: PolicyConfigResourceMatch[]
}

export type PolicyConfigPolicy = {
  id?: string
  name?: string
  description?: string
  parameters?: {[key: string]: GoogleProtobufStruct.Value}
  status?: string
}

export type PolicyConfigConf = {
  parameters?: {[key: string]: GoogleProtobufStruct.Value}
}

export type PolicyConfigObjectSpec = {
  match?: PolicyConfigMatch
  config?: {[key: string]: PolicyConfigConf}
}

export type PolicyConfigObject = {
  metadata?: Metadata
  spec?: PolicyConfigObjectSpec
}

export type EncryptSopsSecretRequest = {
  name?: string
  namespace?: string
  labels?: {[key: string]: string}
  type?: string
  immutable?: boolean
  data?: {[key: string]: string}
  stringData?: {[key: string]: string}
  kustomizationName?: string
  kustomizationNamespace?: string
  clusterName?: string
}

export type EncryptSopsSecretResponse = {
  encryptedSecret?: GoogleProtobufStruct.Value
  path?: string
}

export type ListSopsKustomizationsRequest = {
  clusterName?: string
}

export type ListSopsKustomizationsResponse = {
  kustomizations?: SopsKustomizations[]
  total?: number
}

export type SopsKustomizations = {
  name?: string
  namespace?: string
}

export type SopsSecretMetadata = {
  name?: string
  namespace?: string
  labels?: {[key: string]: string}
}

export type SopsSecret = {
  apiVersion?: string
  kind?: string
  metadata?: SopsSecretMetadata
  data?: {[key: string]: string}
  stringData?: {[key: string]: string}
  type?: string
  immutable?: boolean
  sops?: GoogleProtobufStruct.Value
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
  static RenderAutomation(req: RenderAutomationRequest, initReq?: fm.InitReq): Promise<RenderAutomationResponse> {
    return fm.fetchReq<RenderAutomationRequest, RenderAutomationResponse>(`/v1/enterprise/automations/render`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListGitopsClusters(req: ListGitopsClustersRequest, initReq?: fm.InitReq): Promise<ListGitopsClustersResponse> {
    return fm.fetchReq<ListGitopsClustersRequest, ListGitopsClustersResponse>(`/v1/clusters?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static CreatePullRequest(req: CreatePullRequestRequest, initReq?: fm.InitReq): Promise<CreatePullRequestResponse> {
    return fm.fetchReq<CreatePullRequestRequest, CreatePullRequestResponse>(`/v1/clusters`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static CreateTfControllerPullRequest(req: CreateTfControllerPullRequestRequest, initReq?: fm.InitReq): Promise<CreateTfControllerPullRequestResponse> {
    return fm.fetchReq<CreateTfControllerPullRequestRequest, CreateTfControllerPullRequestResponse>(`/v1/tfcontrollers`, {...initReq, method: "POST", body: JSON.stringify(req)})
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
  static CreateAutomationsPullRequest(req: CreateAutomationsPullRequestRequest, initReq?: fm.InitReq): Promise<CreateAutomationsPullRequestResponse> {
    return fm.fetchReq<CreateAutomationsPullRequestRequest, CreateAutomationsPullRequestResponse>(`/v1/enterprise/automations`, {...initReq, method: "POST", body: JSON.stringify(req)})
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
  static ListEvents(req: ListEventsRequest, initReq?: fm.InitReq): Promise<ListEventsResponse> {
    return fm.fetchReq<ListEventsRequest, ListEventsResponse>(`/v1/enterprise/events?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ListChartsForRepository(req: ListChartsForRepositoryRequest, initReq?: fm.InitReq): Promise<ListChartsForRepositoryResponse> {
    return fm.fetchReq<ListChartsForRepositoryRequest, ListChartsForRepositoryResponse>(`/v1/charts/list?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetValuesForChart(req: GetValuesForChartRequest, initReq?: fm.InitReq): Promise<GetValuesForChartResponse> {
    return fm.fetchReq<GetValuesForChartRequest, GetValuesForChartResponse>(`/v1/charts/values`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetChartsJob(req: GetChartsJobRequest, initReq?: fm.InitReq): Promise<GetChartsJobResponse> {
    return fm.fetchReq<GetChartsJobRequest, GetChartsJobResponse>(`/v1/charts/jobs/${req["jobId"]}?${fm.renderURLSearchParams(req, ["jobId"])}`, {...initReq, method: "GET"})
  }
  static ListWorkspaces(req: ListWorkspacesRequest, initReq?: fm.InitReq): Promise<ListWorkspacesResponse> {
    return fm.fetchReq<ListWorkspacesRequest, ListWorkspacesResponse>(`/v1/workspaces?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetWorkspace(req: GetWorkspaceRequest, initReq?: fm.InitReq): Promise<GetWorkspaceResponse> {
    return fm.fetchReq<GetWorkspaceRequest, GetWorkspaceResponse>(`/v1/workspaces/${req["workspaceName"]}?${fm.renderURLSearchParams(req, ["workspaceName"])}`, {...initReq, method: "GET"})
  }
  static GetWorkspaceRoles(req: GetWorkspaceRequest, initReq?: fm.InitReq): Promise<GetWorkspaceRolesResponse> {
    return fm.fetchReq<GetWorkspaceRequest, GetWorkspaceRolesResponse>(`/v1/workspaces/${req["workspaceName"]}/roles?${fm.renderURLSearchParams(req, ["workspaceName"])}`, {...initReq, method: "GET"})
  }
  static GetWorkspaceRoleBindings(req: GetWorkspaceRequest, initReq?: fm.InitReq): Promise<GetWorkspaceRoleBindingsResponse> {
    return fm.fetchReq<GetWorkspaceRequest, GetWorkspaceRoleBindingsResponse>(`/v1/workspaces/${req["workspaceName"]}/rolebindings?${fm.renderURLSearchParams(req, ["workspaceName"])}`, {...initReq, method: "GET"})
  }
  static GetWorkspaceServiceAccounts(req: GetWorkspaceRequest, initReq?: fm.InitReq): Promise<GetWorkspaceServiceAccountsResponse> {
    return fm.fetchReq<GetWorkspaceRequest, GetWorkspaceServiceAccountsResponse>(`/v1/workspaces/${req["workspaceName"]}/serviceaccounts?${fm.renderURLSearchParams(req, ["workspaceName"])}`, {...initReq, method: "GET"})
  }
  static GetWorkspacePolicies(req: GetWorkspaceRequest, initReq?: fm.InitReq): Promise<GetWorkspacePoliciesResponse> {
    return fm.fetchReq<GetWorkspaceRequest, GetWorkspacePoliciesResponse>(`/v1/workspaces/${req["workspaceName"]}/policies?${fm.renderURLSearchParams(req, ["workspaceName"])}`, {...initReq, method: "GET"})
  }
  static ListExternalSecrets(req: ListExternalSecretsRequest, initReq?: fm.InitReq): Promise<ListExternalSecretsResponse> {
    return fm.fetchReq<ListExternalSecretsRequest, ListExternalSecretsResponse>(`/v1/external-secrets?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetExternalSecret(req: GetExternalSecretRequest, initReq?: fm.InitReq): Promise<GetExternalSecretResponse> {
    return fm.fetchReq<GetExternalSecretRequest, GetExternalSecretResponse>(`/v1/external-secrets/${req["externalSecretName"]}?${fm.renderURLSearchParams(req, ["externalSecretName"])}`, {...initReq, method: "GET"})
  }
  static ListExternalSecretStores(req: ListExternalSecretStoresRequest, initReq?: fm.InitReq): Promise<ListExternalSecretStoresResponse> {
    return fm.fetchReq<ListExternalSecretStoresRequest, ListExternalSecretStoresResponse>(`/v1/external-secrets-stores?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static SyncExternalSecrets(req: SyncExternalSecretsRequest, initReq?: fm.InitReq): Promise<SyncExternalSecretsResponse> {
    return fm.fetchReq<SyncExternalSecretsRequest, SyncExternalSecretsResponse>(`/v1/external-secrets/sync`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListPolicyConfigs(req: ListPolicyConfigsRequest, initReq?: fm.InitReq): Promise<ListPolicyConfigsResponse> {
    return fm.fetchReq<ListPolicyConfigsRequest, ListPolicyConfigsResponse>(`/v1/policy-configs?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetPolicyConfig(req: GetPolicyConfigRequest, initReq?: fm.InitReq): Promise<GetPolicyConfigResponse> {
    return fm.fetchReq<GetPolicyConfigRequest, GetPolicyConfigResponse>(`/v1/policy-configs/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static EncryptSopsSecret(req: EncryptSopsSecretRequest, initReq?: fm.InitReq): Promise<EncryptSopsSecretResponse> {
    return fm.fetchReq<EncryptSopsSecretRequest, EncryptSopsSecretResponse>(`/v1/encrypt-sops-secret`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListSopsKustomizations(req: ListSopsKustomizationsRequest, initReq?: fm.InitReq): Promise<ListSopsKustomizationsResponse> {
    return fm.fetchReq<ListSopsKustomizationsRequest, ListSopsKustomizationsResponse>(`/v1/sops-kustomizations?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
}