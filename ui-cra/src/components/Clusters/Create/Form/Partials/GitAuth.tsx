import React, { FC, Dispatch, useEffect, useState } from 'react';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import {
  CallbackStateContextProvider,
  clearCallbackState,
  getCallbackState,
  getProviderToken,
  GithubDeviceAuthModal,
  RepoInputWithAuth,
} from '@weaveworks/weave-gitops';
import useNotifications from '../../../../../contexts/Notifications';
import useTemplates from '../../../../../contexts/Templates';
import useVersions from '../../../../../contexts/Versions';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';

const GitAuth: FC<{
  showAuthDialog: boolean;
  setShowAuthDialog: Dispatch<React.SetStateAction<boolean>>;
  setEnableCreatePR: Dispatch<React.SetStateAction<boolean>>;
}> = ({ showAuthDialog, setShowAuthDialog, setEnableCreatePR }) => {
  const { setNotifications } = useNotifications();
  const { activeTemplate } = useTemplates();
  const { repositoryURL } = useVersions();

  let initialFormState = {
    url: repositoryURL,
    provider: '',
  };

  const callbackState = getCallbackState();

  if (callbackState) {
    initialFormState = {
      ...initialFormState,
      ...callbackState.state,
    };
    clearCallbackState();
  }

  const [formState, setFormState] = useState(initialFormState);
  const [authSuccess, setAuthSuccess] = useState<boolean>(false);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);

  const credentialsDetected =
    authSuccess ||
    !!getProviderToken(formState.provider as GitProvider) ||
    !!callbackState;

  useEffect(() => {
    setFormState(prevState => ({ ...prevState, url: repositoryURL }));
    setIsAuthenticated(!!formState.url && credentialsDetected);
  }, [repositoryURL, credentialsDetected, formState.url]);

  useEffect(() => {
    if (isAuthenticated) {
      setEnableCreatePR(true);
    }
  }, [isAuthenticated, setEnableCreatePR]);

  return (
    <CallbackStateContextProvider
      callbackState={{
        page: `/clusters` as PageRoute,
        state: formState,
      }}
    >
      <RepoInputWithAuth
        style={{
          marginBottom: weaveTheme.spacing.base,
          width: '80%',
        }}
        isAuthenticated={isAuthenticated}
        onProviderChange={(provider: GitProvider) => {
          setFormState({ ...formState, provider });
        }}
        onChange={e => {
          setFormState({
            ...formState,
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
        value={formState.url}
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
    </CallbackStateContextProvider>
  );
};

export default GitAuth;
