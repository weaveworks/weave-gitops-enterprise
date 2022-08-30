import { createContext, Dispatch, useContext } from 'react';
import { ClusterNamespacedName } from '../../cluster-services/cluster_services.pb';
import {
  GitopsClusterEnriched,
  DeleteClustersPRRequestEnriched,
} from '../../types/custom';

interface ClustersContext {
  clusters: GitopsClusterEnriched[];
  isLoading: boolean;
  count: number | null;
  loading: boolean;
  selectedClusters: ClusterNamespacedName[];
  setSelectedClusters: Dispatch<
    React.SetStateAction<ClusterNamespacedName[] | []>
  >;
  deleteCreatedClusters: (
    data: DeleteClustersPRRequestEnriched,
    token: string,
  ) => Promise<any>;
  getKubeconfig: (
    clusterName: string,
    clusterNamespace: string,
    fileName: string,
  ) => Promise<any>;
  getDashboardAnnotations: (cluster: GitopsClusterEnriched) => {
    [key: string]: string;
  };
  getCluster: (clusterName: string) => GitopsClusterEnriched | null;
  activeCluster: GitopsClusterEnriched | null;
  setActiveCluster: Dispatch<
    React.SetStateAction<GitopsClusterEnriched | null>
  >;
}

export const Clusters = createContext<ClustersContext | null>(null);

export default () => useContext(Clusters) as ClustersContext;
