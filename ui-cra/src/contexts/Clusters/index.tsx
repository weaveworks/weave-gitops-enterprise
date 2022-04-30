import { createContext, Dispatch, useContext } from 'react';
import { GitopsCluster } from '../../capi-server/capi_server.pb';
import { ClusterStatus } from '../../types/kubernetes';

export interface GitopsClusterEnriched extends GitopsCluster {
  status: ClusterStatus;
  pullRequest: {
    type: string;
    url: string;
  };
  type: string;
  updatedAt: string;
  capiCluster: {
    status: string;
  };
}

export interface DeleteClusterPRRequest {
  clusterNames: string[];
  headBranch: string;
  title: string;
  commitMessage: string;
  description: string;
  repositoryUrl?: string;
}
interface ClustersContext {
  clusters: GitopsClusterEnriched[] | [];
  count: number | null;
  disabled: boolean;
  loading: boolean;
  handleRequestSort: (property: string) => void;
  order: string;
  orderBy: string;
  selectedClusters: string[];
  setSelectedClusters: Dispatch<React.SetStateAction<string[] | []>>;
  deleteCreatedClusters: (
    data: DeleteClusterPRRequest,
    token: string,
  ) => Promise<any>;
  deleteConnectedClusters: (clusters: number[]) => void;
  getKubeconfig: (clusterName: string, fileName: string) => void;
}

export const Clusters = createContext<ClustersContext | null>(null);

export default () => useContext(Clusters) as ClustersContext;
