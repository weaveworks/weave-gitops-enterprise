import { createContext } from 'react';
import { ClustersService } from '../../cluster-services/cluster_services.pb';

export type EnterpriseClientContextType = {
  api: typeof ClustersService;
};

export const EnterpriseClientContext =
  createContext<EnterpriseClientContextType>({ api: ClustersService });
