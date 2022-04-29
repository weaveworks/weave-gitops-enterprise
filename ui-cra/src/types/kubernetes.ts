// Note: as of 2019-12-12 the backend will never actually return a status of 'alerts', but it is nonetheless used in fake-eks-cluster-list and therefore is included here in the types since removing it later will be a backwards compatible change.
export type ClusterStatus =
  | 'notConnected'
  | 'ready'
  | 'critical'
  | 'alerting'
  | 'lastSeen'
  | 'pullRequestCreated'
  | 'clusterFound';

export interface NodeGroup {
  desiredCapacity: number;
  instanceType: string;
  name: string;
}

export interface FluxInfo {
  name: string;
  namespace: string;
  repoUrl: string;
  repoBranch: string;
}

export interface GitCommitInfo {
  sha: string;
  author_name: string;
  author_email: string;
  author_date: {
    Time: string;
  };
  message: string;
}

export interface Workspace {
  name: string;
  namespace: string;
}

export interface PullRequest {
  url: string;
  type: string;
}

export interface CAPICluster {
  status: any;
}

export interface Cluster {
  id?: number;
  name: string;
  type?: string;
  token?: string;
  ingressUrl?: string;
  status?: ClusterStatus;
  updatedAt?: string;
  fluxInfo?: FluxInfo[];
  gitCommits?: GitCommitInfo[];
  nodes?: Node[];
  workspaces?: Workspace[];
  pullRequest?: PullRequest;
  capiName?: string;
  capiNamespace?: string;
  capiCluster?: CAPICluster;
}

export interface GitopsCluster {
  name: string;
  namespace: string;
  secretRef: string | null;
  annotations: Object;
  capiClusterRef: { name: string };
  conditions: any[];
  labels: { [key: string]: string };
  capiCluster?: string;
}

export interface Node {
  isControlPlane: boolean;
  kubeletVersion: string;
  name: string;
}

export interface Alert {
  id: number;
  severity: string;
  annotations: { description: string; summary: string; message: string };
  starts_at: string;
  labels: { alertname: string };
  cluster: Cluster;
}

export type ClusterList = Cluster[];
