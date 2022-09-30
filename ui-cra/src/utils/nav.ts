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

export enum NavRoute {
  TerraformObjects = '/terraform_objects',
  TerraformDetail = '/terraform',

  Pipelines = '/pipelines',
  PipelinesDetail = '/pipelins/details',
  Delivery = '/delivery',
  Templates = '/templates',
  Policies = '/policies',
  PoliciesDetail = '/policies/details',

  Clusters = '/clusters',
  Violations = '/clusters/violations',
}

export function getParentNavValue(
  r: NavRoute | V2Routes,
): NavRoute | V2Routes | boolean {
  const [, currentPage] = r.split('/');

  switch (`/${currentPage}`) {
    case V2Routes.Automations:
    case V2Routes.Kustomization:
    case V2Routes.HelmRelease:
      return V2Routes.Automations;

    case V2Routes.Sources:
    case V2Routes.GitRepo:
    case V2Routes.HelmChart:
    case V2Routes.HelmRepo:
    case V2Routes.Bucket:
    case V2Routes.OCIRepository:
      return V2Routes.Sources;

    case V2Routes.FluxRuntime:
      return V2Routes.FluxRuntime;

    case V2Routes.Notifications:
    case V2Routes.Provider:
      return V2Routes.Notifications;

    case NavRoute.Pipelines:
    case NavRoute.PipelinesDetail:
      return NavRoute.Pipelines;

    case NavRoute.TerraformObjects:
    case NavRoute.TerraformDetail:
      return NavRoute.TerraformObjects;

    default:
      return false;
  }
}
