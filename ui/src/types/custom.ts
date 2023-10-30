import {
  CreateDeletionPullRequestRequest,
  GitopsCluster,
  RepositoryChart,
  SourceRef,
  Template,
  TemplateProfile,
} from '../cluster-services/cluster_services.pb';

//
// Utils
//

// Make certain fields of a type required.
// Useful for adapting grpc-gateway-ts fields which are ALL always optional.
// We can mark those we know will always be returned as required
//
type WithRequired<T, K extends keyof T> = T & { [P in K]-?: T[P] };

//
// Types
//

export type Credential = {
  group?: string;
  version?: string;
  kind?: string;
  name?: string;
  namespace?: string;
};

export type Maintainer = {
  name?: string;
  email?: string;
  url?: string;
};

export enum HelmRepositoryType {
  Default = 'Default',
  OCI = 'OCI',
}

export type Interval = {
  hours?: string;
  minutes?: string;
  seconds?: string;
};

export type Condition = {
  type?: string;
  status?: string;
  reason?: string;
  message?: string;
  timestamp?: string;
};

export type HelmRepository = {
  namespace?: string;
  name?: string;
  url?: string;
  interval?: Interval;
  conditions?: Condition[];
  suspended?: boolean;
  lastUpdatedAt?: string;
  clusterName?: string;
  apiVersion?: string;
  repositoryType?: HelmRepositoryType;
  tenant?: string;
};

export type Profile = {
  name: string;
  home?: string;
  sources?: string[];
  description?: string;
  keywords?: string[];
  maintainers?: Maintainer[];
  icon?: string;
  annotations?: { [key: string]: string };
  kubeVersion?: string;
  helmRepository?: HelmRepository;
  availableVersions: string[];
  layer?: string;
};

export type UpdatedProfile = {
  name: Profile['name'];
  repoName: string;
  repoNamespace: string;
  values: { version: string; yaml: string; selected?: boolean }[];
  // optional stuff
  editable?: boolean;
  required: boolean;
  layer?: string;
  namespace?: string;
  selected?: boolean;
};

export interface CAPICluster {
  status: any;
}

export interface GitopsClusterEnriched extends GitopsCluster {
  name: string;
  namespace: string;
  type: string;
  updatedAt: string;
}

export interface TemplateProfilesEnriched extends TemplateProfile {
  name: string;
  sourceRef: WithRequired<SourceRef, 'name' | 'namespace'>;
}

export interface TemplateEnriched extends Template {
  name: string;
  templateKind: string;
  profiles?: TemplateProfilesEnriched[];
}

export type DeleteClustersPRRequestEnriched = WithRequired<
  CreateDeletionPullRequestRequest,
  'headBranch' | 'title' | 'commitMessage' | 'description'
>;

export type EnhancedRepositoryChart = WithRequired<
  RepositoryChart,
  'name' | 'repoName' | 'repoNamespace'
>;

export type PRDefaults = {
  branchName: string;
  pullRequestTitle: string;
  commitMessage: string;
};

export type ProfilesIndex = { [name: string]: UpdatedProfile };

export type RequestError = Error & {
  code?: number;
};
