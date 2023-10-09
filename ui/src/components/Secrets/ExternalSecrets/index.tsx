import {
  Button,
  Flex,
  GitRepository,
  Link,
  Page,
} from '@weaveworks/weave-gitops';
import { useCallback, useContext, useMemo, useState } from 'react';
import { GitProvider } from '../../../api/gitauth/gitauth.pb';
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
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import { EnterpriseClientContext } from '../../../contexts/EnterpriseClient';

const CreateExternalSecret = () => {
  const callbackState = useCallbackState();
  const random = useMemo(() => Math.random().toString(36).substring(7), []);
  const { initialFormData } = getESInitialData(callbackState, random);

  const [showAuthDialog, setShowAuthDialog] = useState(false);

  const [formError, setFormError] = useState<string>('');
  const [validateForm, setValidateForm] = useState<boolean>(false);
  const [formData, setFormData] = useState<ExternalSecret>(initialFormData);
  const handleFormData = (value: any, key: string) => {
    setFormData(f => ({ ...f, [key]: value }));
  };
  const { api } = useContext(EnterpriseClientContext);

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
          const response = await api.CreateAutomationsPullRequest(
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
  }, [api, formData, setNotifications, token, validateToken]);

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
            onSubmit={event => {
              setValidateForm(true);
              validateFormData(event, handleCreateSecret, setFormError);
            }}
          >
            <InputDebounced
              required
              name="secretName"
              label="EXTERNAL SECRET NAME"
              value={formData.secretName}
              handleFormData={val => handleFormData(val, 'secretName')}
              error={validateForm && !formData.secretName}
            />
            <InputDebounced
              required
              name="dataSecretKey"
              label="TARGET K8s SECRET NAME"
              value={formData.dataSecretKey}
              handleFormData={val => handleFormData(val, 'dataSecretKey')}
              error={validateForm && !formData.dataSecretKey}
            />
            <Flex column>
              <ListClusters
                value={formData.clusterName}
                validateForm={validateForm}
                handleFormData={(val: any) => {
                  handleFormData(val, 'clusterName');
                  handleFormData('', 'secretStoreRef');
                }}
              />
            </Flex>
            {formData.clusterName && (
              <ListSecretsStore
                validateForm={validateForm}
                value={formData.secretStore}
                handleFormData={(val: any) => handleSecretStoreChange(val)}
                clusterName={formData.clusterName}
              />
            )}
            {formData.secretStore && (
              <Flex wide>
                <InputDebounced
                  required
                  name="secretStoreType"
                  label="SECRET STORE TYPE"
                  value={formData.secretStoreType}
                  handleFormData={val => {}}
                  disabled={true}
                  error={validateForm && !formData.secretStoreType}
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
                  error={validateForm && !formData.secretNamespace}
                />
              </Flex>
            )}
            <InputDebounced
              required
              name="secretPath"
              label="SECRET PATH"
              value={formData.secretPath}
              handleFormData={val => handleFormData(val, 'secretPath')}
              error={validateForm && !formData.secretPath}
            />
            <SecretProperty
              formData={formData}
              setFormData={setFormData}
              validateForm={validateForm}
            />
            <GitOps
              formData={formData}
              setFormData={setFormData}
              showAuthDialog={showAuthDialog}
              setShowAuthDialog={setShowAuthDialog}
              formError={formError}
              enableGitRepoSelection={true}
            />
            <Flex end className="gitops-cta">
              <Button
                loading={loading}
                type="submit"
                disabled={!isAuthenticated || loading}
              >
                CREATE PULL REQUEST
              </Button>
              <Preview
                formData={formData}
                secretType={SecretType.ES}
                setFormError={setFormError}
              />
            </Flex>
          </FormWrapperSecret>
        </NotificationsWrapper>
      </CallbackStateContextProvider>
    </Page>
  );
};

export default CreateExternalSecret;
