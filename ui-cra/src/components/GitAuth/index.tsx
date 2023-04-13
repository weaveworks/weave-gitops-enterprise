import React, { Dispatch, FC, useEffect, useState } from 'react';
import styled from 'styled-components';
import { GitProvider } from '../../api/gitauth/gitauth.pb';
import { useIsAuthenticated } from '../../hooks/gitprovider';
import { getRepositoryUrl } from '../Templates/Form/utils';
import { GithubDeviceAuthModal } from './GithubDeviceAuthModal';
import { RepoInputWithAuth } from './RepoInputWithAuth';

const RepoInputWithAuthWrapper = styled(RepoInputWithAuth)`
  width: 100%;
  & .auth-message {
    padding-right: ${({ theme }) => theme.spacing.small};
    button {
      min-width: 250px;
    }
    div {
      padding-right: ${({ theme }) => theme.spacing.small};
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
  creatingPR?: boolean;
  setSendPR?: Dispatch<React.SetStateAction<boolean>>;
}> = ({
  formData,
  setFormData,
  showAuthDialog,
  setShowAuthDialog,
  setEnableCreatePR,
  enableGitRepoSelection,
  creatingPR,
  setSendPR,
}) => {
  const [authSuccess, setAuthSuccess] = useState<boolean>(false);
  const { isAuthenticated, loading } = useIsAuthenticated(
    formData.provider,
    creatingPR,
    setSendPR,
  );

  useEffect(() => {
    if (!formData.provider) {
      return;
    }
  }, [formData.provider]);

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
        loading={loading}
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
        value={
          formData?.repo &&
          JSON.stringify({
            value: getRepositoryUrl(formData?.repo),
            key: formData?.repo?.obj?.spec?.url,
          })
        }
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
