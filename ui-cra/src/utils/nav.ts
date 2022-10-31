import { formatURL, Kind, V2Routes } from '@weaveworks/weave-gitops';

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
  Canaries = '/applications/canaries',
  CanaryDetails = '/applications/canary_details',
  Pipelines = '/applications/pipelines',
  PipelineDetails = '/applications/pipelines/details',

  TerraformObjects = '/terraform_objects',
  TerraformDetail = '/terraform',
  Clusters = '/clusters',
  ClusterDashboard = '/cluster',
  DeleteCluster = '/clusters/delete',
  EditCluster = '/clusters/:clusterName/edit',
  PolicyViolations = '/clusters/violations',
  PolicyViolationDetails = '/clusters/violations/details',

  GitlabOauthCallback = '/oauth/gitlab',
  Policies = '/policies',
  PolicyDetails = '/policies/details',

  AddCluster = '/templates/:templateName/create',

  Templates = '/templates',
}

export function reconciledObjectsRoute(s: string, params: any) {
  switch (s) {
    case 'Canary':
      return formatURL(Routes.CanaryDetails, params);

    default:
      return '';
  }
}
