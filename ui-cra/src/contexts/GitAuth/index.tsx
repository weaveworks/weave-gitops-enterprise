import * as React from 'react';
import { useQuery } from 'react-query';
import {
  Applications as applicationsClient,
  // GetGithubAuthStatusResponse,
  // GetGithubDeviceCodeResponse,
} from '../../api/applications/applications.pb';

export interface DialogState {
  open: boolean;
  repoName: string | null;
  success: boolean;
}

export type GitAuthContext = {
  dialogState: DialogState;
  setDialogState: (open: boolean, repoName: string) => void;
  setSuccess: () => void;
  applicationsClient: typeof applicationsClient;
};

export const GitAuth = React.createContext<GitAuthContext>(null as any);

export interface Props {
  children?: any;
}

export const GitAuthProvider: React.FC = ({ children }: Props) => {
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
        applicationsClient,
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
  const { applicationsClient } = React.useContext(GitAuth);
  return useQuery<GetGithubAuthStatusResponse, Error>(
    [GITHUB_AUTH_KEY],
    () => applicationsClient.GetGithubAuthStatus(codeRes),
    { retry: false, refetchInterval: (codeRes.interval || 1) * 1000 },
  );
};

const GITHUB_DEVICE_KEY = 'githubDeviceCode';
export const useGetGithubDeviceCode = () => {
  const { applicationsClient } = React.useContext(GitAuth);
  return useQuery<GetGithubDeviceCodeResponse, Error>(
    [GITHUB_DEVICE_KEY],
    () => applicationsClient.GetGithubDeviceCode({}),
    {
      retry: false,
    },
  );
};

export interface GetGithubDeviceCodeResponse {
  userCode?: string | undefined;
  deviceCode?: string | undefined;
  validationURI?: string | undefined;
  interval?: number | undefined;
}
export interface GetGithubAuthStatusResponse {
  accessToken?: string | undefined;
  error?: string | undefined;
}
