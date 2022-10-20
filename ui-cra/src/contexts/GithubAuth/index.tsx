import React, { useContext } from 'react';
import { useQuery } from 'react-query';
import { applicationsClient } from '@weaveworks/weave-gitops';
import { GetGithubAuthStatusResponse, GetGithubDeviceCodeResponse } from './provider';


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
  return useQuery<GetGithubAuthStatusResponse, Error>(
    [GITHUB_AUTH_KEY],
    () => applicationsClient.GetGithubAuthStatus(codeRes),
    { retry: false, refetchInterval: (codeRes.interval || 1) * 1000 },
  );
};

const GITHUB_DEVICE_KEY = 'githubDeviceCode';
export const useGetGithubDeviceCode = () => {
  return useQuery<GetGithubDeviceCodeResponse, Error>(
    [GITHUB_DEVICE_KEY],
    () => applicationsClient.GetGithubDeviceCode({}),
    {
      retry: false,
    },
  );
};
