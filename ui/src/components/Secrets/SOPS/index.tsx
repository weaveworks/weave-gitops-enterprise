import { MenuItem } from '@material-ui/core';
import { Flex, GitRepository, Link } from '@weaveworks/weave-gitops';
import { useCallback, useMemo, useState } from 'react';
import { GitProvider } from '../../../api/gitauth/gitauth.pb';
import { useAPI } from '../../../contexts/API';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import useNotifications from '../../../contexts/Notifications';
import {
  expiredTokenNotification,
  useIsAuthenticated,
} from '../../../hooks/gitprovider';
import { useCallbackState } from '../../../utils/callback-state';
import { InputDebounced, Select, validateFormData } from '../../../utils/form';
import { Routes } from '../../../utils/nav';
import { removeToken } from '../../../utils/request';
import { clearCallbackState, getProviderToken } from '../../GitAuth/utils';
import { Page } from '../../Layout/App';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import GitOps from '../../Templates/Form/Partials/GitOps';
import { getRepositoryUrl } from '../../Templates/Form/utils';
import ListClusters from '../Shared/ListClusters';
import ListKustomizations from '../Shared/ListKustomizations';
import { Preview } from '../Shared/Preview';
import {
  getFormattedPayload,
  scrollToAlertSection,
  handleError,
  getInitialData,
  SOPS,
  FormWrapperSecret,
} from '../Shared/utils';
import SecretData from './SecretData';

const CreateSOPS = () => {
  const callbackState = useCallbackState();
  const random = useMemo(() => Math.random().toString(36).substring(7), []);
  const { initialFormData } = getInitialData(callbackState, random);

  const [showAuthDialog, setShowAuthDialog] = useState(false);

  const [formError, setFormError] = useState<string>('');
  const [validateForm, setValidateForm] = useState<boolean>(false);
  const [formData, setFormData] = useState<SOPS>(initialFormData);
  const handleFormData = (value: any, key: string) => {
    setFormData(f => ({ ...f, [key]: value }));
  };
  const { setNotifications } = useNotifications();

  const [loading, setLoading] = useState<boolean>(false);
  const token = getProviderToken(formData.provider as GitProvider);

  const { isAuthenticated, validateToken } = useIsAuthenticated(
    formData.provider as GitProvider,
    token,
  );

  const { clustersService } = useAPI();

  const handleCreateSecret = useCallback(() => {
    setLoading(true);

    validateToken()
      .then(async () => {
        try {
          const { encryptionPayload, cluster } = getFormattedPayload(formData);
          const encrypted = await clustersService.EncryptSopsSecret(
            encryptionPayload,
          );
          const response = await clustersService.CreateAutomationsPullRequest(
            {
              headBranch: formData.branchName,
              title: formData.pullRequestTitle,
              description: formData.pullRequestDescription,
              commitMessage: formData.commitMessage,
              repositoryUrl: getRepositoryUrl(formData.repo as GitRepository),
              clusterAutomations: [
                {
                  cluster,
                  isControlPlane: cluster.namespace ? true : false,
                  sopsSecret: {
                    ...encrypted.encryptedSecret,
                  },
                  filePath: encrypted.path,
                },
              ],
            },
            {
              headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
            },
          );
          setNotifications([
            {
              message: {
                component: (
                  <Link href={response.webUrl} newTab>
                    PR created successfully, please review and merge the pull
                    request to apply the changes to the cluster.
                  </Link>
                ),
              },
              severity: 'success',
            },
          ]);
          scrollToAlertSection();
        } catch (error: any) {
          handleError(error, setNotifications);
        } finally {
          setLoading(false);
          removeToken(formData.provider);
          clearCallbackState();
        }
      })
      .catch(() => {
        removeToken(formData.provider);
        setNotifications([expiredTokenNotification]);
      })
      .finally(() => setLoading(false));
  }, [enterprise, formData, setNotifications, token, validateToken]);

  const authRedirectPage = Routes.CreateSopsSecret;

  return (
    <Page
      path={[
        { label: 'Secrets', url: Routes.Secrets },
        { label: 'Create new sops secret' },
      ]}
    >
      <CallbackStateContextProvider
        callbackState={{
          page: authRedirectPage,
          state: {
            formData,
          },
        }}
      >
        <NotificationsWrapper>
          <FormWrapperSecret
            noValidate
            onSubmit={event => {
              setValidateForm(true);
              validateFormData(event, handleCreateSecret, setFormError);
            }}
          >
            <Flex column>
              <ListClusters
                value={formData.clusterName}
                validateForm={validateForm}
                handleFormData={(val: any) => {
                  handleFormData(val, 'clusterName');
                  handleFormData('', 'kustomization');
                }}
              />
              <InputDebounced
                required
                name="secretName"
                label="SECRET NAME"
                value={formData.secretName}
                handleFormData={val => handleFormData(val, 'secretName')}
                error={validateForm && !formData.secretName}
              />
              <InputDebounced
                required
                name="secretNamespace"
                label="SECRET NAMESPACE"
                value={formData.secretNamespace}
                handleFormData={val => handleFormData(val, 'secretNamespace')}
                error={validateForm && !formData.secretNamespace}
              />
            </Flex>
            <h2>Encryption</h2>
            <Select
              required
              name="encryptionType"
              label="ENCRYPT USING"
              value={formData.encryptionType}
              onChange={event =>
                handleFormData(event.target.value, 'encryptionType')
              }
            >
              <MenuItem value="GPG/AGE">GPG / AGE</MenuItem>
            </Select>
            {!!formData.clusterName && (
              <ListKustomizations
                validateForm={validateForm}
                value={formData.kustomization}
                handleFormData={(val: any) =>
                  handleFormData(val, 'kustomization')
                }
                clusterName={formData.clusterName}
              />
            )}
            <h2>Secret Data</h2>
            <p className="secret-data-hint">
              Please note that we will encode the secret values to base64 before
              encryption
            </p>
            <SecretData
              formData={formData}
              setFormData={setFormData}
              validateForm={validateForm}
            />
            <GitOps
              loading={loading}
              isAuthenticated={isAuthenticated}
              formData={formData}
              setFormData={setFormData}
              showAuthDialog={showAuthDialog}
              setShowAuthDialog={setShowAuthDialog}
              formError={formError}
              enableGitRepoSelection={true}
            >
              <Preview formData={formData} setFormError={setFormError} />
            </GitOps>
          </FormWrapperSecret>
        </NotificationsWrapper>
      </CallbackStateContextProvider>
    </Page>
  );
};

export default CreateSOPS;
