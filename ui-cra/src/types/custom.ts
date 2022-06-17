import { GitopsCluster } from '../cluster-services/cluster_services.pb';

export type Credential = {
  group?: string;
  version?: string;
  kind?: string;
  name?: string;
  namespace?: string;
};

export type ListCredentialsResponse = {
  credentials?: Credential[];
  total?: number;
};

export type Maintainer = {
  name?: string;
  email?: string;
  url?: string;
};

export type HelmRepository = {
  name?: string;
  namespace?: string;
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

export type ListProfilesResponse = {
  profiles?: Profile[];
};

export type UpdatedProfile = {
  name: Profile['name'];
  values: { version: string; yaml: string; selected?: boolean }[];
  required: boolean;
  layer?: string;
  namespace?: string;
};

export type ChildrenOccurences = {
  name: string;
  groupVisible: boolean;
  count: number;
};

export interface CAPICluster {
  status: any;
}

export interface GitopsClusterEnriched extends GitopsCluster {
  name: string;
  type: string;
  updatedAt: string;
}

export type ListGitopsClustersResponseEnriched = {
  gitopsClusters: GitopsClusterEnriched[];
  total: number;
};
