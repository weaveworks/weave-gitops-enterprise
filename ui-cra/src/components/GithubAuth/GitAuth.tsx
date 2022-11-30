import React, { FC, Dispatch, useEffect, useState } from 'react';
import styled from 'styled-components';
import { GithubDeviceAuthModal } from '.';
import { GitProvider } from './utils';
import { useIsAuthenticated } from '../../hooks/gitprovider';
import RepoInputWithAuth from './RepoInputWithAuth';

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
        onChange={(e: any) => {
          setFormData({
            ...formData,
            url: e.currentTarget.value,
          });
        }}
        onAuthClick={(provider: GitProvider) => {
          if (provider === ('GitHub' as GitProvider)) {
            setShowAuthDialog(true);
          }
        }}
        required
        id="url"
        label="Source Repo URL"
        variant="standard"
        // this needs to be a dropdown; we need to get the list of git repos that we use when we create apps
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
