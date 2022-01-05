import React, { FC, Dispatch, useEffect, useState } from 'react';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import {
  getProviderToken,
  GithubDeviceAuthModal,
  RepoInputWithAuth,
} from '@weaveworks/weave-gitops';
import useNotifications from '../../../../../contexts/Notifications';
import styled from 'styled-components';

const RepoInputWithAuthWrapper = styled(RepoInputWithAuth)`
  margin-bottom: ${weaveTheme.spacing.base};
  width: 100%;
  & .auth-message {
    width: 300px;
    & .MuiButtonBase-root {
      width: 100%;
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
  const { setNotifications } = useNotifications();

  const [authSuccess, setAuthSuccess] = useState<boolean>(false);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);

  const credentialsDetected =
    authSuccess || !!getProviderToken(formData.provider as GitProvider);

  useEffect(() => {
    setIsAuthenticated(!!formData.url && credentialsDetected);
  }, [credentialsDetected, formData.url]);

  useEffect(() => {
    if (isAuthenticated) {
      setEnableCreatePR(true);
    } else {
      setEnableCreatePR(false);
    }
  }, [isAuthenticated, setEnableCreatePR]);

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
        onClose={() => setShowAuthDialog(false)}
        onSuccess={() => {
          setShowAuthDialog(false);
          setAuthSuccess(true);
          setNotifications([
            {
              message:
                'Authentication completed successfully. Please proceed with creating the PR.',
              variant: 'success',
            },
          ]);
        }}
        open={showAuthDialog}
        repoName="config"
      />
    </>
  );
};

export default GitAuth;
