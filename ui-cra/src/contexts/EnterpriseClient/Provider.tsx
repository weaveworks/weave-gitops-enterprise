import React, { FC, Props } from 'react';
import { EnterpriseClientContext } from './index';
import { UnAuthrizedInterceptor } from '@weaveworks/weave-gitops';
import { ClustersService } from '../../capi-server/capi_server.pb';

const EnterpriseClientProvider: FC = ({ api, children }: Props) => {
  const wrapped = UnAuthrizedInterceptor(ClustersService) as any;
  return (
    <EnterpriseClientContext.Provider value={{ api: wrapped }}>
      {children}
    </EnterpriseClientContext.Provider>
  );
};

export default EnterpriseClientProvider;
