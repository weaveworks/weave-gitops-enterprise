import React from 'react';
import { EnterpriseClientContext } from './index';
import { UnAuthroizedInterceptor } from '@weaveworks/weave-gitops';
import { ClustersService } from '../../capi-server/capi_server.pb';

type Props = {
  api: typeof ClustersService;
  children: any;
};

const EnterpriseClientProvider = ({ api, children }: Props) => {
  const wrapped = UnAuthroizedInterceptor(api) as typeof ClustersService;
  return (
    <EnterpriseClientContext.Provider value={{ api: wrapped }}>
      {children}
    </EnterpriseClientContext.Provider>
  );
};

export default EnterpriseClientProvider;
