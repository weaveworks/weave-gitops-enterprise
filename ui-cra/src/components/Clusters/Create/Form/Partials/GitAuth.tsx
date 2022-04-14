import React, { FC, Dispatch, useEffect, useState } from 'react';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import {
  GithubDeviceAuthModal,
  RepoInputWithAuth,
  theme as weaveTheme,
  useIsAuthenticated,
} from '@weaveworks/weave-gitops';
import styled from 'styled-components';

const RepoInputWithAuthWrapper = styled(RepoInputWithAuth)`
  margin-bottom: ${weaveTheme.spacing.base};
  width: 100%;
  & .auth-message {
    margin-right: ${weaveTheme.spacing.xs};
    button {
      span {
        width: 200px;
      }
    }
  }
`;

const GitAuth: FC<{
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  showAuthDialog: boolean;
  setShowAuthDialog: Dispatch<React.SetStateAction<boolean>>;
  setEnableCreatePR: Dispatch<React.SetStateAction<boolean>>;
}> = ({
  formData,
  setFormData,
  showAuthDialog,
  setShowAuthDialog,
  setEnableCreatePR,
}) => {
  const [authSuccess, setAuthSuccess] = useState<boolean>(false);
  const { isAuthenticated, req: check } = useIsAuthenticated();

  useEffect(() => {
    if (!formData.provider) {
      return;
    }
    check(formData.provider);
  }, [formData.provider, authSuccess]); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (isAuthenticated) {
      setEnableCreatePR(true);
    } else {
      setEnableCreatePR(false);
    }
  }, [authSuccess, isAuthenticated, setEnableCreatePR]);

  return (
    <>
      <RepoInputWithAuthWrapper
        isAuthenticated={isAuthenticated}
        onProviderChange={(provider: GitProvider) => {
          setFormData({ ...formData, provider });
        }}
        onChange={e => {
          setFormData({
            ...formData,
            url: e.currentTarget.value,
          });
        }}
        onAuthClick={provider => {
          if (provider === ('GitHub' as GitProvider)) {
            setShowAuthDialog(true);
          }
        }}
        required
        id="url"
        label="Source Repo URL"
        variant="standard"
        value={formData.url}
        helperText=""
        disabled={true}
      />
      <GithubDeviceAuthModal
        bodyClassName="GithubDeviceAuthModal"
        onClose={() => setShowAuthDialog(false)}
        onSuccess={() => {
          setShowAuthDialog(false);
          setAuthSuccess(true);
        }}
        open={showAuthDialog}
        repoName="config"
      />
    </>
  );
};

export default GitAuth;
