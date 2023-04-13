import React, { Dispatch, FC, useEffect, useState } from 'react';
import styled from 'styled-components';
import { GitProvider } from '../../api/gitauth/gitauth.pb';
import { useIsAuthenticated } from '../../hooks/gitprovider';
import { getRepositoryUrl } from '../Templates/Form/utils';
import { GithubDeviceAuthModal } from './GithubDeviceAuthModal';
import { RepoInputWithAuth } from './RepoInputWithAuth';
import useNotifications from './../../contexts/Notifications';

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
  setCreatingPR?: Dispatch<React.SetStateAction<boolean>>;
  setSendPR?: Dispatch<React.SetStateAction<boolean>>;
}> = ({
  formData,
  setFormData,
  showAuthDialog,
  setShowAuthDialog,
  setEnableCreatePR,
  enableGitRepoSelection,
  creatingPR,
  setCreatingPR,
  setSendPR,
}) => {
  const [authSuccess, setAuthSuccess] = useState<boolean>(false);
  const { isAuthenticated, loading } = useIsAuthenticated(
    formData.provider,
    creatingPR,
  );
  const { setNotifications } = useNotifications();

  useEffect(() => {
    if (!formData.provider) {
      return;
    }
  }, [formData.provider]);

  useEffect(() => {
    // if user is authenticated enable the create PR button
    if (isAuthenticated) {
      setEnableCreatePR(true);
    } else {
      setEnableCreatePR(false);
    }

    // if the user is not authenticated, checking of the token has finished and the user has already clicked the Create PR button
    // show notification informing user they need to authenticate again as the token has expired
    if (!isAuthenticated && creatingPR) {
      setCreatingPR && setCreatingPR(false);
      setNotifications([
        {
          message: {
            text: 'Your token seems to have expired. Please go through the authentication process again and then submit your create PR request.',
          },
          severity: 'error',
          display: 'bottom',
        },
      ]);
      return;
    }

    // if user is authenticated and in the process of sending the PR, send the PR data to the api
    /// sendPR will trigger validateFormData and then submit in form
    if (isAuthenticated && creatingPR) {
      console.log('isAuthenticated', isAuthenticated, 'creatingPR', creatingPR);
      setSendPR && setSendPR(true);
    }
  }, [
    authSuccess,
    isAuthenticated,
    setEnableCreatePR,
    loading,
    setNotifications,
    creatingPR,
    setCreatingPR,
    setSendPR,
  ]);

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
