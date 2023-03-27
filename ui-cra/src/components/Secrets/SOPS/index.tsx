import { CircularProgress, MenuItem } from '@material-ui/core';
import { Button, GitRepository, Link } from '@weaveworks/weave-gitops';
import { useCallback, useMemo, useState } from 'react';
import { GitProvider } from '../../../api/gitauth/gitauth.pb';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import useNotifications from '../../../contexts/Notifications';
import { useCallbackState } from '../../../utils/callback-state';
import { Select, validateFormData } from '../../../utils/form';
import { Routes } from '../../../utils/nav';
import { removeToken } from '../../../utils/request';
import {
  CreateDeploymentObjects,
  encryptSopsSecret,
} from '../../Applications/utils';
import { clearCallbackState, getProviderToken } from '../../GitAuth/utils';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import GitOps from '../../Templates/Form/Partials/GitOps';
import { getRepositoryUrl } from '../../Templates/Form/utils';
import InputDebounced from './InputDebounced';
import ListClusters from './ListClusters';
import ListKustomizations from './ListKustomizations';
import { PreviewModal } from './PreviewModal';
import SecretData from './SecretData';
import { FormWrapper } from './styles';
import {
  getFormattedPayload,
  scrollToAlertSection,
  handleError,
  getInitialData,
  SOPS,
} from './utils';

const CreateSOPS = () => {
  const callbackState = useCallbackState();
  const random = useMemo(() => Math.random().toString(36).substring(7), []);
  const { initialFormData } = getInitialData(callbackState, random);

  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const [enableCreatePR, setEnableCreatePR] = useState<boolean>(false);

  const [formError, setFormError] = useState<string>('');
  const [formData, setFormData] = useState<SOPS>(initialFormData);
  const handleFormData = (value: any, key: string) => {
    setFormData(f => ({ ...f, [key]: value }));
  };
  const { setNotifications } = useNotifications();

  const [loading, setLoading] = useState<boolean>(false);
  const handleCreateSecret = useCallback(async () => {
    setLoading(true);

    try {
      const { encryptionPayload, cluster } = getFormattedPayload(formData);
      const encrypted = await encryptSopsSecret(encryptionPayload);
      const response = await CreateDeploymentObjects(
        {
          head_branch: formData.branchName,
          title: formData.pullRequestTitle,
          description: formData.pullRequestDescription,
          commitMessage: formData.commitMessage,
          repositoryUrl: getRepositoryUrl(formData.repo as GitRepository),
          clusterAutomations: [
            {
              cluster,
              isControlPlane: cluster.namespace ? true : false,
              sops_secret: {
                ...encrypted.encryptedSecret,
              },
              file_path: encrypted.path,
            },
          ],
        },
        getProviderToken(formData.provider as GitProvider),
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
  }, [formData, setNotifications]);

  const authRedirectPage = Routes.CreateSopsSecret;

  return (
    <PageTemplate
      documentTitle="SOPS"
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
        <ContentWrapper>
          <FormWrapper
            noValidate
            onSubmit={event =>
              validateFormData(event, handleCreateSecret, setFormError)
            }
          >
            <div className="group-section">
              <div className="form-group">
                <ListClusters
                  value={formData.clusterName}
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
                />
                <InputDebounced
                  required
                  name="secretNamespace"
                  label="SECRET NAMESPACE"
                  value={formData.secretNamespace}
                  handleFormData={val => handleFormData(val, 'secretNamespace')}
                />
              </div>
            </div>
            <div className="group-section">
              <h2>Encryption</h2>
              <div className="form-group">
                <Select
                  className="form-section"
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
                    value={formData.kustomization}
                    handleFormData={(val: any) =>
                      handleFormData(val, 'kustomization')
                    }
                    clusterName={formData.clusterName}
                  />
                )}
              </div>
            </div>
            <div className="group-section">
              <h2>Secret Data</h2>
              <p className="secret-data-hint">
                Please note that we will encode the secret values to base64
                before encryption
              </p>
              <SecretData formData={formData} setFormData={setFormData} />

              <PreviewModal formData={formData} />
            </div>
            <GitOps
              formData={formData}
              setFormData={setFormData}
              showAuthDialog={showAuthDialog}
              setShowAuthDialog={setShowAuthDialog}
              setEnableCreatePR={setEnableCreatePR}
              formError={formError}
              enableGitRepoSelection={true}
            />

            <div className="create-cta">
              <Button type="submit" disabled={!enableCreatePR || loading}>
                CREATE PULL REQUEST
                {loading && (
                  <CircularProgress
                    size={'1rem'}
                    style={{ marginLeft: '4px' }}
                  />
                )}
              </Button>
            </div>
          </FormWrapper>
        </ContentWrapper>
      </CallbackStateContextProvider>
    </PageTemplate>
  );
};

export default CreateSOPS;
