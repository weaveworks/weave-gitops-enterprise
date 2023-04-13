import { poller } from '@weaveworks/weave-gitops';
import { useState, useContext, useEffect, Dispatch } from 'react';
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
import useNotifications, {
  NotificationData,
} from './../contexts/Notifications';

const providerTokenHeaderName = 'Git-Provider-Token';

const expiredTokenNotification = {
  message: {
    text: 'Your token seems to have expired. Please go through the authentication process again and then submit your create PR request.',
  },
  severity: 'error',
  display: 'bottom',
} as NotificationData;

export function useIsAuthenticated(
  provider: GitProvider,
  creatingPR?: boolean,
  setSendPR?: Dispatch<React.SetStateAction<boolean>>,
) {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>();
  const [loading, setLoading] = useState<boolean>(false);
  const { gitAuthClient } = useContext(GitAuth);
  const token = getProviderToken(provider);
  const { setNotifications } = useNotifications();

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
        setSendPR && setSendPR(true);
      } else {
        setIsAuthenticated(false);
        if (creatingPR) {
          setNotifications([expiredTokenNotification]);
        }
        return;
      }
    };

    validateProviderToken()
      .catch(() => {
        setIsAuthenticated(false);
        if (creatingPR) {
          setNotifications([expiredTokenNotification]);
          return;
        }
      })
      .finally(() => setLoading(false));
  }, [gitAuthClient, provider, creatingPR, token, setNotifications, setSendPR]);

  return {
    isAuthenticated,
    loading,
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
