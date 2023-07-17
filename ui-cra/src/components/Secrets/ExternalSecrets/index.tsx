import {
  Button,
  Flex,
  GitRepository,
  Link,
  LoadingPage,
  Page,
} from '@weaveworks/weave-gitops';
import { useCallback, useMemo, useState } from 'react';
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
import { createDeploymentObjects } from '../../Applications/utils';
import { clearCallbackState, getProviderToken } from '../../GitAuth/utils';
import GitOps from '../../Templates/Form/Partials/GitOps';
import { FormWrapper, getRepositoryUrl } from '../../Templates/Form/utils';
import ListClusters from '../Shared/ListClusters';
import { PreviewModal, SecretType } from '../Shared/PreviewModal';
import {
  ExternalSecret,
  getESFormattedPayload,
  getESInitialData,
  handleError,
  scrollToAlertSection,
} from '../Shared/utils';
import ListSecretsStore from './ListSecretsStore';
import { SecretProperty } from './SecretProperty';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import styled from 'styled-components';

export const FormWrapperSecret = styled(FormWrapper)`
  .secret-data-list {
    display: flex;
    align-items: self-start;
    width: 100%;
    .remove-icon {
      margin-top: 25px;
      color: ${props => props.theme.colors.neutral30};
      cursor: pointer;
    }
  }
  .secret-data-hint {
    background-color: ${props => props.theme.colors.primaryLight05};
    padding: ${props => props.theme.spacing.xs};
    font-weight: 600;
    width: fit-content;
    border-radius: 4px;
    margin-top: 0px;
  }
  .add-secret-data {
    margin-bottom: ${props => props.theme.spacing.medium};
  }
`;

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
          const response = await createDeploymentObjects(
            {
              head_branch: formData.branchName,
              title: formData.pullRequestTitle,
              description: formData.pullRequestDescription,
              commitMessage: formData.commitMessage,
              repositoryUrl: getRepositoryUrl(formData.repo as GitRepository),
              clusterAutomations: [payload],
            },
            token,
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
  }, [formData, setNotifications, token, validateToken]);

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
            <PreviewModal formData={formData} secretType={SecretType.ES} />
            <GitOps
              formData={formData}
              setFormData={setFormData}
              showAuthDialog={showAuthDialog}
              setShowAuthDialog={setShowAuthDialog}
              formError={formError}
              enableGitRepoSelection={true}
            />
            {loading ? (
              <LoadingPage className="create-loading" />
            ) : (
              <Flex end className="create-cta">
                <Button type="submit" disabled={!isAuthenticated}>
                  CREATE PULL REQUEST
                </Button>
              </Flex>
            )}
          </FormWrapperSecret>
        </NotificationsWrapper>
      </CallbackStateContextProvider>
    </Page>
  );
};

export default CreateExternalSecret;
