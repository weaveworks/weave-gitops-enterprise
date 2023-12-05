import * as React from 'react';
import { useQuery } from 'react-query';
import {
  GetGithubAuthStatusResponse,
  GetGithubDeviceCodeResponse,
  ParseRepoURLResponse,
} from '../../api/gitauth/gitauth.pb';
import { useEnterpriseClient } from '../API';

export interface DialogState {
  open: boolean;
  repoName: string | null;
  success: boolean;
}

export const useGitAuth = () => {
  const { gitAuth } = useEnterpriseClient();
  const [dialogState, setDialogState] = React.useState<DialogState>({
    open: false,
    repoName: null,
    success: false,
  });

  return {
    gitAuth,
    dialogState,
    setDialogState: (open: boolean, repoName: string) =>
      setDialogState({ ...dialogState, open, repoName }),
    setSuccess: () => setDialogState({ ...dialogState, success: true }),
  };
};

const GITHUB_AUTH_KEY = 'githubAuth';
export const useGetGithubAuthStatus = (
  codeRes: GetGithubDeviceCodeResponse,
) => {
  const { gitAuth } = useGitAuth();
  return useQuery<GetGithubAuthStatusResponse, Error>(
    [GITHUB_AUTH_KEY],
    () => gitAuth.GetGithubAuthStatus(codeRes),
    { retry: false, refetchInterval: (codeRes.interval || 1) * 1000 },
  );
};

const GITHUB_DEVICE_KEY = 'githubDeviceCode';
export const useGetGithubDeviceCode = () => {
  const { gitAuth } = useGitAuth();
  return useQuery<GetGithubDeviceCodeResponse, Error>(
    [GITHUB_DEVICE_KEY],
    () => gitAuth.GetGithubDeviceCode({}),
    {
      retry: false,
    },
  );
};

export const useParseRepoUrl = (value: string) => {
  const { gitAuth } = useGitAuth();
  return useQuery<ParseRepoURLResponse, Error>(
    [value],
    () =>
      gitAuth.ParseRepoURL({
        url: value,
      }),
    {
      retry: false,
    },
  );
};
