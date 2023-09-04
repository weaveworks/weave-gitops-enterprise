import { poller } from '@weaveworks/weave-gitops';
import { useState, useContext, useEffect, useCallback } from 'react';
import {
  GetGithubAuthStatusResponse,
  GetGithubDeviceCodeResponse,
  GitProvider,
} from '../api/gitauth/gitauth.pb';
import {
  getProviderToken,
  GrpcErrorCodes,
  storeProviderToken,
} from '../components/GitAuth/utils';
import { GitAuth } from '../contexts/GitAuth';
import { NotificationData } from './../contexts/Notifications';

const providerTokenHeaderName = 'Git-Provider-Token';

export const expiredTokenNotification = {
  message: {
    text: 'Your token seems to have expired. Please go through the authentication process again and then submit your create PR request.',
  },
  severity: 'error',
  display: 'bottom',
} as NotificationData;

export function useIsAuthenticated(
  provider: GitProvider,
  token: string | null,
) {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>();
  const [loading, setLoading] = useState<boolean>(false);
  const { gitAuthClient } = useContext(GitAuth);

  const validateToken = useCallback(() => {
    const makeHeaders = () =>
      new Headers({
        [providerTokenHeaderName]: `token ${token}`,
      });
    const headers = makeHeaders();

    return gitAuthClient.ValidateProviderToken({ provider }, { headers });
  }, [gitAuthClient, provider, token]);

  useEffect(() => {
    if (provider === ('' as GitProvider)) {
      return;
    }
    if (token) {
      setLoading(true);
      validateToken()
        .then(res => setIsAuthenticated(res?.valid ? true : false))
        .catch(() => setIsAuthenticated(false))
        .finally(() => setLoading(false));
    } else {
      setIsAuthenticated(false);
    }
  }, [validateToken, token, provider]);

  return {
    validateToken,
    isAuthenticated,
    loading,
  };
}
