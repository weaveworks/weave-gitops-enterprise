import {
  GitAuth as gitAuthClient,
  GetGithubAuthStatusResponse,
  GetGithubDeviceCodeResponse,
  ParseRepoURLResponse,
} from '../../api/gitauth/gitauth.pb';
import * as React from 'react';
import { useQuery } from 'react-query';

export interface DialogState {
  open: boolean;
  repoName: string | null;
  success: boolean;
}

export type GitAuthContext = {
  dialogState: DialogState;
  setDialogState: (open: boolean, repoName: string) => void;
  setSuccess: () => void;
  gitAuthClient: typeof gitAuthClient;
};

export const GitAuth = React.createContext<GitAuthContext>(null as any);

export interface Props {
  children?: any;
  api?: typeof gitAuthClient;
}

export const GitAuthProvider: React.FC = ({ children, api }: Props) => {
  const [dialogState, setDialogState] = React.useState<DialogState>({
    open: false,
    repoName: null,
    success: false,
  });

  return (
    <GitAuth.Provider
      value={{
        dialogState,
        setDialogState: (open: boolean, repoName: string) =>
          setDialogState({ ...dialogState, open, repoName }),
        setSuccess: () => setDialogState({ ...dialogState, success: true }),
        gitAuthClient: api || gitAuthClient,
      }}
    >
      {children}
    </GitAuth.Provider>
  );
};

const GITHUB_AUTH_KEY = 'githubAuth';
export const useGetGithubAuthStatus = (
  codeRes: GetGithubDeviceCodeResponse,
) => {
  const { gitAuthClient } = React.useContext(GitAuth);
  return useQuery<GetGithubAuthStatusResponse, Error>(
    [GITHUB_AUTH_KEY],
    () => gitAuthClient.GetGithubAuthStatus(codeRes),
    { retry: false, refetchInterval: (codeRes.interval || 1) * 1000 },
  );
};

const GITHUB_DEVICE_KEY = 'githubDeviceCode';
export const useGetGithubDeviceCode = () => {
  const { gitAuthClient } = React.useContext(GitAuth);
  return useQuery<GetGithubDeviceCodeResponse, Error>(
    [GITHUB_DEVICE_KEY],
    () => gitAuthClient.GetGithubDeviceCode({}),
    {
      retry: false,
    },
  );
};

export const useParseRepoUrl = (value: string) => {
  const { gitAuthClient } = React.useContext(GitAuth);
  return useQuery<ParseRepoURLResponse, Error>(
    [value],
    () =>
      gitAuthClient.ParseRepoURL({
        url: value,
      }),
    {
      retry: false,
    },
  );
};
