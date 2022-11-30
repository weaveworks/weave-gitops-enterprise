import React, { useContext } from 'react';
import { useQuery } from 'react-query';
import { applicationsClient } from '@weaveworks/weave-gitops';
import {
  GetGithubAuthStatusResponse,
  GetGithubDeviceCodeResponse,
} from '../../components/GithubAuth/utils';

interface Props {
  api: typeof applicationsClient;
  children: any;
}

export const GithubAuthContext = React.createContext<typeof applicationsClient>(
  null as any,
);

export const GithubAuthProvider = ({ api, children }: Props) => (
  <GithubAuthContext.Provider value={api}>
    {children}
  </GithubAuthContext.Provider>
);

export const useGithubAuth = () => useContext(GithubAuthContext);

const GITHUB_AUTH_KEY = 'githubAuth';
export const useGetGithubAuthStatus = (
  codeRes: GetGithubDeviceCodeResponse,
) => {
  const githubService = useGithubAuth();
  return useQuery<GetGithubAuthStatusResponse, Error>(
    [GITHUB_AUTH_KEY],
    () => githubService.GetGithubAuthStatus(codeRes),
    { retry: false, refetchInterval: (codeRes.interval || 1) * 1000 },
  );
};

const GITHUB_DEVICE_KEY = 'githubDeviceCode';
export const useGetGithubDeviceCode = () => {
  const githubService = useGithubAuth();
  return useQuery<GetGithubDeviceCodeResponse, Error>(
    [GITHUB_DEVICE_KEY],
    () => githubService.GetGithubDeviceCode({}),
    {
      retry: false,
    },
  );
};
