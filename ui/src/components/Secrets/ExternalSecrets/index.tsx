import { Flex, GitRepository, Link, Page } from '@weaveworks/weave-gitops';
import { useCallback, useMemo, useState } from 'react';
import { GitProvider } from '../../../api/gitauth/gitauth.pb';
import { useEnterpriseClient } from '../../../contexts/API';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import useNotifications from '../../../contexts/Notifications';
import {
  expiredTokenNotification,
  useIsAuthenticated,
} from '../../../hooks/gitprovider';
import { useCallbackState } from '../../../utils/callback-state';
import { InputDebounced, validateFormData } from '../../../utils/form';
import { Routes } from '../../../utils/nav';
import { removeToken } from '../../../utils/request';
import { clearCallbackState, getProviderToken } from '../../GitAuth/utils';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import GitOps from '../../Templates/Form/Partials/GitOps';
import { getRepositoryUrl } from '../../Templates/Form/utils';
import ListClusters from '../Shared/ListClusters';
import { Preview, SecretType } from '../Shared/Preview';
import {
  ExternalSecret,
  FormWrapperSecret,
  getESFormattedPayload,
  getESInitialData,
  handleError,
  scrollToAlertSection,
} from '../Shared/utils';
import ListSecretsStore from './ListSecretsStore';
import { SecretProperty } from './SecretProperty';

const CreateExternalSecret = () => {
  const callbackState = useCallbackState();
  const random = useMemo(() => Math.random().toString(36).substring(7), []);
  const { initialFormData } = getESInitialData(callbackState, random);

  const [showAuthDialog, setShowAuthDialog] = useState(false);

  const [formError, setFormError] = useState<string>('');
  const [formData, setFormData] = useState<ExternalSecret>(initialFormData);
  const handleFormData = (value: any, key: string) => {
    setFormData(f => ({ ...f, [key]: value }));
  };
  const { clustersService } = useEnterpriseClient();

  const handleSecretStoreChange = (value: string) => {
    const [secretStoreRef, secretStoreKind, secretNamespace, secretStoreType] =
      value.split('/');

    setFormData(f => ({
      ...f,
      secretStore: value,
      secretStoreRef: secretStoreRef || '',
      secretStoreKind: secretStoreKind || '',
      secretNamespace: secretNamespace || '',
      secretStoreType: secretStoreType || '',
      defaultSecretNamespace: secretNamespace || '',
    }));
  };
  const { setNotifications } = useNotifications();

  const [loading, setLoading] = useState<boolean>(false);
  const token = getProviderToken(formData.provider as GitProvider);

  const { isAuthenticated, validateToken } = useIsAuthenticated(
    formData.provider as GitProvider,
    token,
  );

  const handleCreateSecret = useCallback(() => {
    setLoading(true);

    validateToken()
      .then(async () => {
        try {
          const payload = getESFormattedPayload(formData);
          const response = await clustersService.CreateAutomationsPullRequest(
            {
              headBranch: formData.branchName,
              title: formData.pullRequestTitle,
              description: formData.pullRequestDescription,
              commitMessage: formData.commitMessage,
              repositoryUrl: getRepositoryUrl(formData.repo as GitRepository),
              clusterAutomations: [payload],
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
  }, [clustersService, formData, setNotifications, token, validateToken]);

  const authRedirectPage = Routes.CreateSecret;

  return (
    <Page
      path={[
        { label: 'Secrets', url: Routes.Secrets },
        { label: 'Create new external secret' },
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
            onSubmit={event =>
              validateFormData(event, handleCreateSecret, setFormError)
            }
          >
            <InputDebounced
              required
              name="secretName"
              label="EXTERNAL SECRET NAME"
              value={formData.secretName}
              handleFormData={val => handleFormData(val, 'secretName')}
              error={formError === 'secretName' && !formData.secretName}
            />
            <InputDebounced
              required
              name="dataSecretKey"
              label="TARGET K8s SECRET NAME"
              value={formData.dataSecretKey}
              handleFormData={val => handleFormData(val, 'dataSecretKey')}
              error={formError === 'dataSecretKey' && !formData.dataSecretKey}
            />
            <Flex column>
              <ListClusters
                value={formData.clusterName}
                handleFormData={(val: any) => {
                  handleFormData(val, 'clusterName');
                  handleFormData('', 'secretStoreRef');
                }}
                error={formError === 'clusterName' && !formData.clusterName}
              />
            </Flex>
            {formData.clusterName && (
              <ListSecretsStore
                // value={formData.secretStore}
                value="test"
                handleFormData={(val: any) => handleSecretStoreChange(val)}
                clusterName={formData.clusterName}
                error={formError === 'secretStoreRef' && !formData.secretStore}
              />
            )}
            {formData.secretStore && (
              <Flex wide>
                <InputDebounced
                  required
                  name="secretStoreType"
                  label="SECRET STORE TYPE"
                  value={formData.secretStoreType}
                  handleFormData={val => {
                    return;
                  }}
                  disabled={true}
                  error={
                    formError === 'secretStoreType' && !formData.secretStoreType
                  }
                />
                <InputDebounced
                  required
                  name="secretNamespace"
                  label="SECRET NAMESPACE"
                  value={formData.secretNamespace}
                  handleFormData={val => handleFormData(val, 'secretNamespace')}
                  disabled={
                    !!formData.secretNamespace &&
                    formData.defaultSecretNamespace === formData.secretNamespace
                  }
                  error={
                    formError === 'secretNamespace' && !formData.secretNamespace
                  }
                />
              </Flex>
            )}
            <InputDebounced
              required
              name="secretPath"
              label="SECRET PATH"
              value={formData.secretPath}
              handleFormData={val => handleFormData(val, 'secretPath')}
              error={formError === 'secretPath' && !formData.secretPath}
            />
            <SecretProperty
              formData={formData}
              setFormData={setFormData}
              formError={formError}
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
              <Preview
                formData={formData}
                secretType={SecretType.ES}
                setFormError={setFormError}
              />
            </GitOps>
          </FormWrapperSecret>
        </NotificationsWrapper>
      </CallbackStateContextProvider>
    </Page>
  );
};

export default CreateExternalSecret;
