import React from 'react';
import { EnterpriseClientContext } from './index';
import qs from 'query-string';

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

function UnAuthorizedInterceptor(api: any): typeof ClustersService {
  const wrapped: any = {};
  //   Wrap each API method in a check that redirects to the signin page if a 401 is returned.
  for (const method of Object.getOwnPropertyNames(api)) {
    if (typeof api[method] != 'function') {
      continue;
    }
    wrapped[method] = (req: any, initReq: any) => {
      return api[method](req as any, initReq as any).catch((err: any) => {
        if (err.code === 401) {
          window.location.replace(
            AuthRoutes.AUTH_PATH_SIGNIN +
              '?' +
              qs.stringify({
                // eslint-disable-next-line no-restricted-globals
                redirect: location.pathname + location.search,
              }),
          );
        }
        throw err;
      });
    };
  }
  return wrapped as typeof ClustersService;
}

const EnterpriseClientProvider = ({ api, children }: Props) => {
  const wrapped = UnAuthorizedInterceptor(api);
  return (
    <EnterpriseClientContext.Provider value={{ api: wrapped }}>
      {children}
    </EnterpriseClientContext.Provider>
  );
};

export default EnterpriseClientProvider;
