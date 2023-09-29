import React from 'react';
import { ClustersService } from '../../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from './index';

type Props = {
  api: typeof ClustersService;
  children: any;
};

const EnterpriseClientProvider = ({ api, children }: Props) => {
  return (
    <EnterpriseClientContext.Provider value={{ api }}>
      {children}
    </EnterpriseClientContext.Provider>
  );
};

export default EnterpriseClientProvider;
