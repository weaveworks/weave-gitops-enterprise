import * as React from 'react';

type Props = {
  children?: any;
};

interface DiaglogState {
  open: boolean;
  repoName: string | null;
  success: boolean;
}

export interface GithubAuthContextType {
  dialogState: DiaglogState;
  setDialogState: (open: boolean, repoName: string) => void;
  setSuccess: () => void;
}

export const GithubAuthContext =
  React.createContext<GithubAuthContextType | null>(null);

export default function GithubAuthContextProvider({ children }: Props) {
  const [dialogState, setDialogState] = React.useState<DiaglogState>({
    open: false,
    repoName: null,
    success: false,
  });

  const value: GithubAuthContextType = {
    dialogState,
    setDialogState: (open: boolean, repoName: string) =>
      setDialogState({ ...dialogState, open, repoName }),
    setSuccess: () => setDialogState({ ...dialogState, success: true }),
  };
  return <GithubAuthContext.Provider value={value} children={children} />;
}
