import { poller, useRequestState } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { useCallback, useState, useContext } from 'react';
import {
  GetGithubAuthStatusResponse,
  GetGithubDeviceCodeResponse,
  GitProvider,
  ValidateProviderTokenResponse,
} from '../api/gitauth/gitauth.pb';
import {
  getProviderToken,
  GrpcErrorCodes,
  makeHeaders,
  storeProviderToken,
} from '../components/GitAuth/utils';
import { GitAuth } from '../contexts/GitAuth';

export function useIsAuthenticated() {
  const [res, loading, error, req] =
    useRequestState<ValidateProviderTokenResponse>();
  const { gitAuthClient } = useContext(GitAuth);

  // const [isAuthenticated, setIsAuthenticated] = useState(
  //   error ? false : res?.valid,
  // );

  // console.log(isAuthenticated);

  return {
    isAuthenticated: error ? false : res?.valid,
    // setIsAuthenticated,
    loading,
    error,
    req: useCallback(
      async (provider: GitProvider) => {
        //@ts-ignore
        const headers = makeHeaders(_.bind(getProviderToken, this, provider));
        await req(
          gitAuthClient.ValidateProviderToken({ provider }, { headers }),
        );
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
