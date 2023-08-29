import React from 'react';
import { EnterpriseClientContext } from './index';
import { ClustersService } from '../../cluster-services/cluster_services.pb';

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
  return (
    <EnterpriseClientContext.Provider value={{ api }}>
      {children}
    </EnterpriseClientContext.Provider>
  );
};

export default EnterpriseClientProvider;
