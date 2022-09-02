import React from 'react';
import { EnterpriseClientContext } from './index';
import { ClustersService } from '../../cluster-services/cluster_services.pb';
import { UnAuthorizedInterceptor } from '@weaveworks/weave-gitops';
import { Core } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';

type Props = {
  api: typeof ClustersService;
  children: any;
};

export enum AuthRoutes {
  USER_INFO = '/oauth2/userinfo',
  SIGN_IN = '/oauth2/sign_in',
  LOG_OUT = '/oauth2/logout',
  AUTH_PATH_SIGNIN = '/sign_in',
}

const EnterpriseClientProvider = ({ api, children }: Props) => {
  const wrapped = UnAuthorizedInterceptor(api as unknown as typeof Core);
  return (
    <EnterpriseClientContext.Provider value={{ api: wrapped as unknown as typeof ClustersService }}>
      {children}
    </EnterpriseClientContext.Provider>
  );
};

export default EnterpriseClientProvider;
