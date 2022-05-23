import { createContext, useContext } from 'react';
import { ClustersService } from '../../cluster-services/clusters_grpc_pb';

export type EnterpriseClientContextType = {
  api: typeof ClustersService;
};

export const EnterpriseClientContext =
  createContext<EnterpriseClientContextType>({ api: ClustersService });

export default () =>
  useContext(EnterpriseClientContext) as EnterpriseClientContextType;
