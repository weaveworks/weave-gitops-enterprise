import { poller, useRequestState } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { useCallback, useState, useContext } from 'react';
import {
  GetGithubAuthStatusResponse,
  GetGithubDeviceCodeResponse,
  GitProvider,
  ValidateProviderTokenResponse,
} from '../api/applications/applications.pb';
import {
  getProviderToken,
  GrpcErrorCodes,
  makeHeaders,
  storeProviderToken,
} from '../components/GithubAuth/utils';
import { GitAuth } from '../contexts/GitAuth';

export function useIsAuthenticated() {
  const [res, loading, error, req] =
    useRequestState<ValidateProviderTokenResponse>();
  const { applicationsClient } = useContext(GitAuth);

  return {
    isAuthenticated: error ? false : res?.valid,
    loading,
    error,
    req: useCallback(
      (provider: GitProvider) => {
        //@ts-ignore
        const headers = makeHeaders(_.bind(getProviderToken, this, provider));
        req(
          applicationsClient.ValidateProviderToken({ provider }, { headers }),
        );
      },
      // eslint-disable-next-line react-hooks/exhaustive-deps
      [applicationsClient],
    ),
  };
}

export default function useAuth() {
  const [loading, setLoading] = useState(true);
  const { applicationsClient } = useContext(GitAuth);

  const getGithubDeviceCode = () => {
    setLoading(true);
    return applicationsClient
      .GetGithubDeviceCode({})
      .finally(() => setLoading(false));
  };

  const getGithubAuthStatus = (codeRes: GetGithubDeviceCodeResponse) => {
    let poll: any;
    return {
      cancel: () => clearInterval(poll),
      promise: new Promise<GetGithubAuthStatusResponse>((accept, reject) => {
        poll = poller(() => {
          applicationsClient
            .GetGithubAuthStatus(codeRes)
            .then(res => {
              clearInterval(poll);
              accept(res);
            })
            .catch(({ code, message }) => {
              // Unauthenticated means we can keep polling.
              //  On anything else, stop polling and report.
              if (code !== GrpcErrorCodes.Unauthenticated) {
                clearInterval(poll);
                reject({ message });
              }
            });
        }, codeRes?.interval && (codeRes?.interval + 1) * 1000);
      }),
    };
  };

  return {
    loading,
    getGithubDeviceCode,
    getGithubAuthStatus,
    getProviderToken,
    storeProviderToken,
  };
}
