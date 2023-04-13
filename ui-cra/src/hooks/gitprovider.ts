import { poller, useRequestState } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { useCallback, useState, useContext, useEffect } from 'react';
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

const providerTokenHeaderName = 'Git-Provider-Token';

export function useIsAuthenticated(
  provider: GitProvider,
  creatingPR?: boolean,
) {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>();
  const [loading, setLoading] = useState<boolean>(false);
  const { gitAuthClient } = useContext(GitAuth);
  const token = getProviderToken(provider);

  useEffect(() => {
    const makeHeaders = () =>
      new Headers({
        [providerTokenHeaderName]: `token ${token}`,
      });
    const headers = makeHeaders();

    setLoading(true);

    const validateProviderToken = async () => {
      const res = await gitAuthClient.ValidateProviderToken(
        { provider },
        { headers },
      );
      if (res?.valid) {
        setIsAuthenticated(true);
      } else {
        setIsAuthenticated(false);
      }
    };

    validateProviderToken()
      .catch(error => {
        console.log('this is an error', error);
        setIsAuthenticated(false);
      })
      .finally(() => setLoading(false));
  }, [gitAuthClient, provider, creatingPR, token]);

  return {
    isAuthenticated,
    loading,
    // error,
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
