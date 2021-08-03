import { createContext, Dispatch, useContext } from 'react';
import { Cluster } from '../../types/kubernetes';

interface ClustersContext {
  clusters: Cluster[] | [];
  count: number | null;
  disabled: boolean;
  loading: boolean;
  handleRequestSort: (property: string) => void;
  handleSetPageParams: (page: number, perPage: number) => void;
  order: string;
  orderBy: string;
  selectedClusters: string[];
  setSelectedClusters: Dispatch<React.SetStateAction<string[] | []>>;
  deleteCreatedClusters: (clusters: string[]) => void;
  deleteConnectedClusters: (clusters: number[]) => void;
  getKubeconfig: (clusterName: string, fileName: string) => void;
}

export const Clusters = createContext<ClustersContext | null>(null);

export default () => useContext(Clusters) as ClustersContext;
