import { poller, useRequestState } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { useCallback, useState, useContext, useEffect } from 'react';
import {
  GetGithubAuthStatusResponse,
  GetGithubDeviceCodeResponse,
  GitProvider,
  ParseRepoURLResponse,
  ValidateProviderTokenResponse,
} from '../api/gitauth/gitauth.pb';
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
  const { gitAuthClient } = useContext(GitAuth);

  return {
    isAuthenticated: error ? false : res?.valid,
    loading,
    error,
    req: useCallback(
      (provider: GitProvider) => {
        //@ts-ignore
        const headers = makeHeaders(_.bind(getProviderToken, this, provider));
        req(gitAuthClient.ValidateProviderToken({ provider }, { headers }));
      },
      // eslint-disable-next-line react-hooks/exhaustive-deps
      [gitAuthClient],
    ),
  };
}

export default function useAuth() {
  const [loading, setLoading] = useState(true);
  const { gitAuthClient } = useContext(GitAuth);

  const getGithubDeviceCode = () => {
    setLoading(true);
    return gitAuthClient
      .GetGithubDeviceCode({})
      .finally(() => setLoading(false));
  };

  const getGithubAuthStatus = (codeRes: GetGithubDeviceCodeResponse) => {
    let poll: any;
    return {
      cancel: () => clearInterval(poll),
      promise: new Promise<GetGithubAuthStatusResponse>((accept, reject) => {
        poll = poller(() => {
          gitAuthClient
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
