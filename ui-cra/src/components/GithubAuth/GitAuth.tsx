import React, { FC, Dispatch, useEffect, useState } from 'react';
import {
  RepoInputWithAuth,
  useIsAuthenticated,
} from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { GithubDeviceAuthModal } from '.';
import { GitProvider } from '../../contexts/GithubAuth/provider';

const RepoInputWithAuthWrapper = styled(RepoInputWithAuth)`
  margin-bottom: ${({ theme }) => theme.spacing.medium};
  width: 100%;
  & .auth-message {
    margin-top: ${({ theme }) => theme.spacing.base};
    button,
    > div {
      width: 200px;
      margin-right: ${({ theme }) => theme.spacing.medium};
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
  }, [formData.provider, authSuccess, check]);

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
      {showAuthDialog && (
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
      )}
    </>
  );
};

export default GitAuth;
