import React from 'react';
import { useAPI } from '../API';
import { EnterpriseClientContext } from './index';
import { ClustersService } from '../../cluster-services/cluster_services.pb';

type Props = {
  api?: typeof ClustersService;
  children: any;
};

const EnterpriseClientProvider = ({ api: apiOverride, children }: Props) => {
  const { enterprise } = useAPI();
  return (
    <EnterpriseClientContext.Provider
      value={{ api: apiOverride || enterprise }}
    >
      {children}
    </EnterpriseClientContext.Provider>
  );
};

export default EnterpriseClientProvider;
