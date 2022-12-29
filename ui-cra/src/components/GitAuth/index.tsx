import React, { FC, Dispatch, useEffect, useState } from 'react';
import styled from 'styled-components';
import { GithubDeviceAuthModal } from './GithubDeviceAuthModal';
import { GitProvider } from '../../api/gitauth/gitauth.pb';
import { useIsAuthenticated } from '../../hooks/gitprovider';
import RepoInputWithAuth from './RepoInputWithAuth';

const RepoInputWithAuthWrapper = styled(RepoInputWithAuth)`
  width: 100%;
  & .auth-message {
    margin-top: ${({ theme }) => theme.spacing.base};
    margin-left: 0;
    margin-right: ${({ theme }) => theme.spacing.small};
    button {
      min-width: 250px;
    }
  }
`;

const GitAuth: FC<{
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  showAuthDialog: boolean;
  setShowAuthDialog: Dispatch<React.SetStateAction<boolean>>;
  setEnableCreatePR: Dispatch<React.SetStateAction<boolean>>;
  enableGitRepoSelection?: boolean;
}> = ({
  formData,
  setFormData,
  showAuthDialog,
  setShowAuthDialog,
  setEnableCreatePR,
  enableGitRepoSelection,
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
        onAuthClick={(provider: GitProvider) => {
          if (provider === ('GitHub' as GitProvider)) {
            setShowAuthDialog(true);
          }
        }}
        required
        id="url"
        label="Source Repo URL"
        variant="standard"
        value={formData?.repo?.obj?.spec?.url}
        description=""
        formData={formData}
        setFormData={setFormData}
        enableGitRepoSelection={enableGitRepoSelection}
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
