import { createContext, Dispatch, useContext } from 'react';
import { GitopsCluster } from '../../capi-server/capi_server.pb';

export interface GitopsClusterEnriched extends GitopsCluster {
  type: string;
  updatedAt: string;
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
  isLoading: boolean;
  count: number | null;
  loading: boolean;
  selectedClusters: string[];
  setSelectedClusters: Dispatch<React.SetStateAction<string[] | []>>;
  deleteCreatedClusters: (
    data: DeleteClusterPRRequest,
    token: string,
  ) => Promise<any>;
  getKubeconfig: (clusterName: string, fileName: string) => void;
}

export const Clusters = createContext<ClustersContext | null>(null);

export default () => useContext(Clusters) as ClustersContext;
