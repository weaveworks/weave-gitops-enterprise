import React from 'react';
import { EnterpriseClientContext } from './index';
import { UnAuthorizedInterceptor } from '@weaveworks/weave-gitops';
import { ClustersService } from '../../capi-server/capi_server.pb';

type Props = {
  api: typeof ClustersService;
  children: any;
};

const EnterpriseClientProvider = ({ api, children }: Props) => {
  const wrapped = UnAuthorizedInterceptor(api) as typeof ClustersService;
  return (
    <EnterpriseClientContext.Provider value={{ api: wrapped }}>
      {children}
    </EnterpriseClientContext.Provider>
  );
};

export default EnterpriseClientProvider;
