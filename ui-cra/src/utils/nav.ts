import { Kind, V2Routes } from '@weaveworks/weave-gitops';

export function getKindRoute(k: Kind | string): string {
  switch (k) {
    case Kind.GitRepository:
    case 'GitRepository':
      return V2Routes.GitRepo;

    case Kind.Bucket:
    case 'Bucket':
      return V2Routes.Bucket;

    case Kind.HelmRepository:
    case 'HelmRepository':
      return V2Routes.HelmRepo;

    case Kind.HelmChart:
    case 'HelmChart':
      return V2Routes.HelmChart;

    case Kind.Kustomization:
    case 'Kustomization':
      return V2Routes.Kustomization;

    case Kind.HelmRelease:
    case 'HelmRelease':
      return V2Routes.HelmRelease;

    default:
      return V2Routes.NotImplemented;
  }
}

export enum Routes {
  Applications = '/applications',
  AddApplication = '/applications/create',
  Canaries = '/delivery',
  CanaryDetails = '/canary_details',
  Pipelines = '/pipelines',
  PipelineDetails = '/pipelines/details',

  GitOpsRun = '/gitopsrun',
  GitOpsRunDetail = '/gitopsrun/detail',

  TerraformObjects = '/terraform',
  TerraformDetail = '/terraform/object',
  Clusters = '/clusters',
  ClusterDashboard = '/cluster',
  DeleteCluster = '/clusters/delete',
  EditResource = '/resources/edit',

  PolicyViolations = '/clusters/violations',
  PolicyViolationDetails = '/violations/details',

  GitlabOauthCallback = '/oauth/gitlab',
  BitBucketOauthCallback = '/oauth/bitbucketserver',
  AzureDevOpsOauthCallback = '/oauth/azuredevops',
  Policies = '/policies',
  PolicyDetails = '/policies/details',

  AddCluster = '/templates/:templateName/create',

  Templates = '/templates',

  Workspaces = '/workspaces',
  WorkspaceDetails = '/workspaces/details',

  Secrets = '/secrets',
  SecretDetails = '/secrets/details',
  CreateSecret = '/secrets/create',
  CreateSopsSecret = '/secrets/sops/create',

  PolicyConfigs = '/policyConfigs',
  PolicyConfigsDetails = '/policyConfigs/details',
  CreatePolicyConfig = '/policyConfigs/create',


  GitOpsSets = '/gitopssets',
  GitOpsSetDetail = '/gitopssets/object',

  ImageAutomation = '/image_automation',

  Explorer = '/explorer',
  ExplorerAccessRules = '/explorer/access_rules',
  Notifications = '/notifications',
}
