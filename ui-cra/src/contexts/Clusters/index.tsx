import { createContext, Dispatch, useContext } from 'react';
import { GitopsCluster } from '../../types/kubernetes';

export interface DeleteClusterPRRequest {
  clusterNames: string[];
  headBranch: string;
  title: string;
  commitMessage: string;
  description: string;
  repositoryUrl?: string;
}
interface ClustersContext {
  clusters: GitopsCluster[] | [];
  count: number | null;
  loading: boolean;
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
