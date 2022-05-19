import { createContext, useContext } from 'react';
import { ClustersService } from '../../capi-server/capi_server.pb';

export type EnterpriseClientContextType = {
  api: typeof ClustersService;
};

export const EnterpriseClientContext =
  createContext<EnterpriseClientContextType>({ api: ClustersService });

export default () =>
  useContext(EnterpriseClientContext) as EnterpriseClientContextType;
