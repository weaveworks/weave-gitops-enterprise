import { ClustersService } from '../../cluster-services/cluster_services.pb';
import { createContext } from 'react';

export type EnterpriseClientContextType = {
  api: typeof ClustersService;
};

export const EnterpriseClientContext =
  createContext<EnterpriseClientContextType>({ api: ClustersService });
