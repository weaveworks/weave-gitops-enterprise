import { createContext, Dispatch, useContext } from 'react';
import { Cluster } from '../../types/kubernetes';

export interface DeleteClusterPRRequest {
  clusterNames: string[];
  headBranch: string;
  title: string;
  commitMessage: string;
  description: string;
  repositoryUrl?: string;
}
interface ClustersContext {
  clusters: Cluster[] | [];
  count: number | null;
  disabled: boolean;
  loading: boolean;
  creatingPR: boolean;
  handleRequestSort: (property: string) => void;
  handleSetPageParams: (page: number, perPage: number) => void;
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
