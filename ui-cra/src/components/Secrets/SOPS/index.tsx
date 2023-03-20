import { MenuItem } from '@material-ui/core';
import { useCallback, useMemo, useState } from 'react';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import { Select, validateFormData } from '../../../utils/form';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import GitOps from '../../Templates/Form/Partials/GitOps';
import { useCallbackState } from '../../../utils/callback-state';
import InputDebounced from './InputDebounced';
import SecretData from './SecretData';
import { FormWrapper } from './styles';
import ListClusters from './ListClusters';
import ListKustomizations from './ListKustomizations';
import { Button, LoadingPage } from '@weaveworks/weave-gitops';

export enum SecretDataType {
  value,
  KeyValue,
}
export interface SOPS {
  clusterName: string;
  secretName: string;
  secretNamespace: string;
  encryptionType: string;
  kustomization: string;
  secretData: { key: string; value: string }[];
  secretValue: string;
  repo: string | null;
  provider: string;
  branchName: string;
  pullRequestTitle: string;
  commitMessage: string;
  pullRequestDescription: string;
  secretType: SecretDataType;
}
function getInitialData(
  callbackState: { state: { formData: SOPS } } | null,
  random: string,
) {
  let defaultFormData = {
    repo: null,
    provider: '',
    branchName: `add-SOPS-secret-branch-${random}`,
    pullRequestTitle: 'Add SOPS Secret',
    commitMessage: 'Add SOPS Secret',
    pullRequestDescription: 'This PR adds a new SOPS Secret',
    clusterName: '',
    secretName: '',
    secretNamespace: '',
    encryptionType: 'GPG/AGE',
    kustomization: '',
    secretData: [{ key: '', value: '' }],
    secretValue: '',
    secretType: SecretDataType.value,
  };

  const initialFormData = {
    ...defaultFormData,
    ...callbackState?.state?.formData,
  };

  return { initialFormData };
}

const CreateSOPS = () => {
  const callbackState = useCallbackState();
  const random = useMemo(() => Math.random().toString(36).substring(7), []);
  const { initialFormData } = getInitialData(callbackState, random);

  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const [enableCreatePR, setEnableCreatePR] = useState<boolean>(false);

  const [formError, setFormError] = useState<string>('');
  const [formData, setFormData] = useState<SOPS>(initialFormData);
  const handleCreateSecret = useCallback(() => {}, []);
  const handleFormData = (value: any, key: string) => {
    setFormData(f => (f = { ...f, [key]: value }));
  };

  const [loading, setLoading] = useState<boolean>(false);

  const authRedirectPage = `/secrets/create`;

  return (
    <PageTemplate
      documentTitle="SOPS"
      path={[
        { label: 'Secrets', url: Routes.Secrets },
        { label: 'Create new SOPS' },
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
                  handleFormData={(val: any) =>
                    handleFormData(val, 'clusterName')
                  }
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

            <SecretData formData={formData} handleFormData={handleFormData} />
            <GitOps
              formData={formData}
              setFormData={setFormData}
              showAuthDialog={showAuthDialog}
              setShowAuthDialog={setShowAuthDialog}
              setEnableCreatePR={setEnableCreatePR}
              formError={formError}
              enableGitRepoSelection={true}
            />

            {loading ? (
              <LoadingPage className="create-loading" />
            ) : (
              <div className="create-cta">
                <Button type="submit" disabled={!enableCreatePR}>
                  CREATE PULL REQUEST
                </Button>
              </div>
            )}
          </FormWrapper>
        </ContentWrapper>
      </CallbackStateContextProvider>
    </PageTemplate>
  );
};

export default CreateSOPS;
