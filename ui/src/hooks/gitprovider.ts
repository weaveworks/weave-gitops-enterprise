import { useState, useEffect, useCallback } from 'react';
import { GitProvider } from '../api/gitauth/gitauth.pb';
import { useEnterpriseClient } from '../contexts/API';
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
  const { gitAuth } = useEnterpriseClient();

  const validateToken = useCallback(() => {
    const makeHeaders = () =>
      new Headers({
        [providerTokenHeaderName]: `token ${token}`,
      });
    const headers = makeHeaders();

    return gitAuth.ValidateProviderToken({ provider }, { headers });
  }, [gitAuth, provider, token]);

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
