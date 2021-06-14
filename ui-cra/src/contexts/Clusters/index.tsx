import { createContext, useContext } from 'react';
import { Cluster } from '../../types/kubernetes';

interface ClustersContext {
  clusters: Cluster[] | [];
  error: string | null;
  count: number | null;
  disabled: boolean;
  loading: boolean;
  handleRequestSort: (property: string) => void;
  handleSetPageParams: (page: number, perPage: number) => void;
  order: string;
  orderBy: string;
  addCluster: (data: any) => void;
}

export const Clusters = createContext<ClustersContext | null>(null);

export default () => useContext(Clusters) as ClustersContext;
