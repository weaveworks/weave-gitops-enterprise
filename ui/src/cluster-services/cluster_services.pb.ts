/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "./fetch.pb"
import * as GoogleApiHttpbody from "./google/api/httpbody.pb"
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
  name?: string
  templateKind?: string
  namespace?: string
}

export type GetTemplateResponse = {
  template?: Template
}

export type ListTemplateParamsRequest = {
  name?: string
  templateKind?: string
  namespace?: string
}

export type ListTemplateParamsResponse = {
  parameters?: Parameter[]
  objects?: TemplateObject[]
}

export type ListTemplateProfilesRequest = {
  name?: string
  templateKind?: string
  namespace?: string
}

export type ListTemplateProfilesResponse = {
  profiles?: TemplateProfile[]
  objects?: TemplateObject[]
}

export type RenderTemplateRequest = {
  name?: string
  values?: {[key: string]: string}
  credentials?: Credential
  templateKind?: string
  clusterNamespace?: string
  profiles?: ProfileValues[]
  kustomizations?: Kustomization[]
  namespace?: string
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
  renderedTemplates?: CommitFile[]
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

export type CreatePullRequestRequest = {
  repositoryUrl?: string
  headBranch?: string
  baseBranch?: string
  title?: string
  description?: string
  name?: string
  parameterValues?: {[key: string]: string}
  commitMessage?: string
  credentials?: Credential
  values?: ProfileValues[]
  repositoryApiUrl?: string
  clusterNamespace?: string
  kustomizations?: Kustomization[]
  namespace?: string
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
  name?: string
  parameterValues?: {[key: string]: string}
  commitMessage?: string
  repositoryApiUrl?: string
  namespace?: string
}

export type CreateTfControllerPullRequestResponse = {
  webUrl?: string
}

export type ClusterNamespacedName = {
  namespace?: string
  name?: string
}

export type CreateDeletionPullRequestRequest = {
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

export type CreateDeletionPullRequestResponse = {
  webUrl?: string
}

export type ListCredentialsRequest = {
}

export type ListCredentialsResponse = {
  credentials?: Credential[]
  total?: number
}

export type GetKubeconfigRequest = {
  name?: string
  namespace?: string
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
  secretStoreRef?: ExternalSecretStoreRef
  target?: ExternalSecretTarget
  data?: ExternalSecretData[]
  dataFrom?: ExternalSecretDataFromRemoteRef
}

export type ExternalSecretStoreRef = {
  name?: string
  kind?: string
}

export type ExternalSecretTarget = {
  name?: string
}

export type ExternalSecretData = {
  secretKey?: string
  remoteRef?: ExternalSecretRemoteRef
}

export type ExternalSecretRemoteRef = {
  key?: string
  property?: string
}

export type ExternalSecretDataFromRemoteRef = {
  extract?: ExternalSecretDataRemoteRef
}

export type ExternalSecretDataRemoteRef = {
  key?: string
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
  repositoryUrl?: string
  managementClusterName?: string
  uiConfig?: string
  gitHostTypes?: {[key: string]: string}
}

export type PolicyParamRepeatedString = {
  values?: string[]
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
  kind?: string
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
  kind?: string
}

export type WorkspaceServiceAccount = {
  name?: string
  namespace?: string
  timestamp?: string
  manifest?: string
  kind?: string
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
  name?: string
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
  properties?: {[key: string]: string}
  version?: string
  status?: string
  timestamp?: string
  yaml?: string
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
    return fm.fetchReq<GetTemplateRequest, GetTemplateResponse>(`/v1/namespaces/${req["namespace"]}/templates/${req["name"]}?${fm.renderURLSearchParams(req, ["namespace", "name"])}`, {...initReq, method: "GET"})
  }
  static ListTemplateParams(req: ListTemplateParamsRequest, initReq?: fm.InitReq): Promise<ListTemplateParamsResponse> {
    return fm.fetchReq<ListTemplateParamsRequest, ListTemplateParamsResponse>(`/v1/namespaces/${req["namespace"]}/templates/${req["name"]}/params?${fm.renderURLSearchParams(req, ["namespace", "name"])}`, {...initReq, method: "GET"})
  }
  static ListTemplateProfiles(req: ListTemplateProfilesRequest, initReq?: fm.InitReq): Promise<ListTemplateProfilesResponse> {
    return fm.fetchReq<ListTemplateProfilesRequest, ListTemplateProfilesResponse>(`/v1/namespaces/${req["namespace"]}/templates/${req["name"]}/profiles?${fm.renderURLSearchParams(req, ["namespace", "name"])}`, {...initReq, method: "GET"})
  }
  static RenderTemplate(req: RenderTemplateRequest, initReq?: fm.InitReq): Promise<RenderTemplateResponse> {
    return fm.fetchReq<RenderTemplateRequest, RenderTemplateResponse>(`/v1/namespaces/${req["namespace"]}/templates/${req["name"]}/render`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static CreatePullRequest(req: CreatePullRequestRequest, initReq?: fm.InitReq): Promise<CreatePullRequestResponse> {
    return fm.fetchReq<CreatePullRequestRequest, CreatePullRequestResponse>(`/v1/namespaces/${req["namespace"]}/templates/${req["name"]}/pull-request`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static CreateDeletionPullRequest(req: CreateDeletionPullRequestRequest, initReq?: fm.InitReq): Promise<CreateDeletionPullRequestResponse> {
    return fm.fetchReq<CreateDeletionPullRequestRequest, CreateDeletionPullRequestResponse>(`/v1/templates/deletion-pull-request`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static RenderAutomation(req: RenderAutomationRequest, initReq?: fm.InitReq): Promise<RenderAutomationResponse> {
    return fm.fetchReq<RenderAutomationRequest, RenderAutomationResponse>(`/v1/automations/render`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static CreateAutomationsPullRequest(req: CreateAutomationsPullRequestRequest, initReq?: fm.InitReq): Promise<CreateAutomationsPullRequestResponse> {
    return fm.fetchReq<CreateAutomationsPullRequestRequest, CreateAutomationsPullRequestResponse>(`/v1/automations/pull-request`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListCredentials(req: ListCredentialsRequest, initReq?: fm.InitReq): Promise<ListCredentialsResponse> {
    return fm.fetchReq<ListCredentialsRequest, ListCredentialsResponse>(`/v1/templates/capi-identities?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static CreateTfControllerPullRequest(req: CreateTfControllerPullRequestRequest, initReq?: fm.InitReq): Promise<CreateTfControllerPullRequestResponse> {
    return fm.fetchReq<CreateTfControllerPullRequestRequest, CreateTfControllerPullRequestResponse>(`/v1/tfcontrollers/pull-request`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListGitopsClusters(req: ListGitopsClustersRequest, initReq?: fm.InitReq): Promise<ListGitopsClustersResponse> {
    return fm.fetchReq<ListGitopsClustersRequest, ListGitopsClustersResponse>(`/v1/clusters?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetKubeconfig(req: GetKubeconfigRequest, initReq?: fm.InitReq): Promise<GoogleApiHttpbody.HttpBody> {
    return fm.fetchReq<GetKubeconfigRequest, GoogleApiHttpbody.HttpBody>(`/v1/namespaces/${req["namespace"]}/clusters/${req["name"]}/kubeconfig?${fm.renderURLSearchParams(req, ["namespace", "name"])}`, {...initReq, method: "GET"})
  }
  static GetEnterpriseVersion(req: GetEnterpriseVersionRequest, initReq?: fm.InitReq): Promise<GetEnterpriseVersionResponse> {
    return fm.fetchReq<GetEnterpriseVersionRequest, GetEnterpriseVersionResponse>(`/v1/enterprise/version?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetConfig(req: GetConfigRequest, initReq?: fm.InitReq): Promise<GetConfigResponse> {
    return fm.fetchReq<GetConfigRequest, GetConfigResponse>(`/v1/config?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
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
    return fm.fetchReq<GetWorkspaceRequest, GetWorkspaceResponse>(`/v1/workspaces/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static GetWorkspaceRoles(req: GetWorkspaceRequest, initReq?: fm.InitReq): Promise<GetWorkspaceRolesResponse> {
    return fm.fetchReq<GetWorkspaceRequest, GetWorkspaceRolesResponse>(`/v1/workspaces/${req["name"]}/roles?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static GetWorkspaceRoleBindings(req: GetWorkspaceRequest, initReq?: fm.InitReq): Promise<GetWorkspaceRoleBindingsResponse> {
    return fm.fetchReq<GetWorkspaceRequest, GetWorkspaceRoleBindingsResponse>(`/v1/workspaces/${req["name"]}/rolebindings?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static GetWorkspaceServiceAccounts(req: GetWorkspaceRequest, initReq?: fm.InitReq): Promise<GetWorkspaceServiceAccountsResponse> {
    return fm.fetchReq<GetWorkspaceRequest, GetWorkspaceServiceAccountsResponse>(`/v1/workspaces/${req["name"]}/serviceaccounts?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static GetWorkspacePolicies(req: GetWorkspaceRequest, initReq?: fm.InitReq): Promise<GetWorkspacePoliciesResponse> {
    return fm.fetchReq<GetWorkspaceRequest, GetWorkspacePoliciesResponse>(`/v1/workspaces/${req["name"]}/policies?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static ListExternalSecrets(req: ListExternalSecretsRequest, initReq?: fm.InitReq): Promise<ListExternalSecretsResponse> {
    return fm.fetchReq<ListExternalSecretsRequest, ListExternalSecretsResponse>(`/v1/external-secrets?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetExternalSecret(req: GetExternalSecretRequest, initReq?: fm.InitReq): Promise<GetExternalSecretResponse> {
    return fm.fetchReq<GetExternalSecretRequest, GetExternalSecretResponse>(`/v1/namespaces/${req["namespace"]}/external-secrets/${req["externalSecretName"]}?${fm.renderURLSearchParams(req, ["namespace", "externalSecretName"])}`, {...initReq, method: "GET"})
  }
  static ListExternalSecretStores(req: ListExternalSecretStoresRequest, initReq?: fm.InitReq): Promise<ListExternalSecretStoresResponse> {
    return fm.fetchReq<ListExternalSecretStoresRequest, ListExternalSecretStoresResponse>(`/v1/external-secrets-stores?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static SyncExternalSecrets(req: SyncExternalSecretsRequest, initReq?: fm.InitReq): Promise<SyncExternalSecretsResponse> {
    return fm.fetchReq<SyncExternalSecretsRequest, SyncExternalSecretsResponse>(`/v1/external-secrets/sync`, {...initReq, method: "PATCH", body: JSON.stringify(req)})
  }
  static EncryptSopsSecret(req: EncryptSopsSecretRequest, initReq?: fm.InitReq): Promise<EncryptSopsSecretResponse> {
    return fm.fetchReq<EncryptSopsSecretRequest, EncryptSopsSecretResponse>(`/v1/encrypt-sops-secret`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListSopsKustomizations(req: ListSopsKustomizationsRequest, initReq?: fm.InitReq): Promise<ListSopsKustomizationsResponse> {
    return fm.fetchReq<ListSopsKustomizationsRequest, ListSopsKustomizationsResponse>(`/v1/sops-kustomizations?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ListPolicyConfigs(req: ListPolicyConfigsRequest, initReq?: fm.InitReq): Promise<ListPolicyConfigsResponse> {
    return fm.fetchReq<ListPolicyConfigsRequest, ListPolicyConfigsResponse>(`/v1/policy-configs?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetPolicyConfig(req: GetPolicyConfigRequest, initReq?: fm.InitReq): Promise<GetPolicyConfigResponse> {
    return fm.fetchReq<GetPolicyConfigRequest, GetPolicyConfigResponse>(`/v1/policy-configs/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
}