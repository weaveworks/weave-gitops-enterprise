import { createContext, Dispatch, useContext } from 'react';
import { GitopsClusterEnriched } from '../../types/custom';

export interface DeleteClusterPRRequest {
  clusterNames: string[];
  headBranch: string;
  title: string;
  commitMessage: string;
  description: string;
  repositoryUrl?: string;
}
interface ClustersContext {
  clusters: GitopsClusterEnriched[];
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
  getDashboardAnnotations: (cluster: GitopsClusterEnriched) => {
    [key: string]: string;
  };
  getCluster: (clusterName: string) => GitopsClusterEnriched | null;
}

export const Clusters = createContext<ClustersContext | null>(null);

export default () => useContext(Clusters) as ClustersContext;
