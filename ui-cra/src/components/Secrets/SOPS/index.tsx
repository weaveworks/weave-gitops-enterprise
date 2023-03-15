import { MenuItem } from '@material-ui/core';
import {
  Kind,
  Kustomization,
  theme,
  useListAutomations,
} from '@weaveworks/weave-gitops';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import { useCallback, useMemo, useState } from 'react';
import styled from 'styled-components';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import { useListImageObjects } from '../../../contexts/ImageAutomation';
import { useListCluster } from '../../../hooks/clusters';
import { useListObjects } from '../../../hooks/listObjects';
import { Select, Input, validateFormData } from '../../../utils/form';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import GitOps from '../../Templates/Form/Partials/GitOps';
interface SOPS {
  clusterName: string;
  secretName: string;
  secretNamespace: string;
  encryptionType: string;
  kustomization: string;
  secretData: { [key: string]: string } | string;
  repo: string;
  provider: string;
  branchName: string;
  pullRequestTitle: string;
  commitMessage: string;
  pullRequestDescription: string;
}

const { medium } = theme.spacing;
const { neutral20, neutral10 } = theme.colors;

const FormWrapper = styled.form`
  .group-section {
    width: 100%;
    border-bottom: 1px dashed ${neutral20};
    .form-group {
      display: flex;
      flex-direction: column;
    }
    .form-section {
      width: 40%;
      .Mui-disabled {
        background: ${neutral10} !important;
        border-color: ${neutral20} !important;
      }
      .MuiInputBase-root {
        margin-right: ${medium};
      }
    }
  }
`;

const CreateSOPS = () => {
  const random = useMemo(() => Math.random().toString(36).substring(7), []);
  let { isLoading, data } = useListCluster();
  const { data: kustomizations } = useListObjects(
    Kustomization,
    Kind.Kustomization,
    '',
    { retry: false },
  );
  const [loading, setLoading] = useState<boolean>(false);
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const [enableCreatePR, setEnableCreatePR] = useState<boolean>(false);

  const [formError, setFormError] = useState<string>('');
  const [formData, setFormData] = useState<SOPS>({
    repo: '',
    provider: '',
    branchName: `add-external-secret-branch-${random}`,
    pullRequestTitle: 'Add External SOPS Secret',
    commitMessage: 'Add External SOPS Secret',
    pullRequestDescription: 'This PR adds a new SOPS Secret',
    clusterName: '',
    secretName: '',
    secretNamespace: '',
    encryptionType: '',
    kustomization: '',
    secretData: '',
  });
  const handleCreateSecret = useCallback(() => {}, []);
  const handleFormData = (event: React.ChangeEvent<any>, key: string) => {
    const value = event.target.value;
    setFormData(f => (f = { ...f, [key]: value }));
  };
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
        <ContentWrapper loading={isLoading}>
          {data?.gitopsClusters && (
            <FormWrapper
              noValidate
              onSubmit={event =>
                validateFormData(event, handleCreateSecret, setFormError)
              }
            >
              <div className="group-section">
                <div className="form-group">
                  <Select
                    className="form-section"
                    name="clusterName"
                    required
                    label="TARGET CLUSTER"
                    onChange={event => handleFormData(event, 'clusterName')}
                    value={formData.clusterName}
                  >
                    {data?.gitopsClusters?.map((option, index: number) => {
                      return (
                        <MenuItem key={index} value={option.name}>
                          {option.name}
                        </MenuItem>
                      );
                    })}
                  </Select>
                  <Input
                    className="form-section"
                    required
                    name="secretName"
                    label="EXTERNAL SECRET NAME"
                    value={formData.secretName}
                    onChange={event => handleFormData(event, 'secretName')}
                  />
                  <Input
                    className="form-section"
                    required
                    name="dataSecretKey"
                    label="TARGET K8s SECRET NAME"
                    value={formData.secretNamespace}
                    onChange={event => handleFormData(event, 'secretNamespace')}
                  />
                </div>
              </div>
              <div className="group-section">
                <h4>Encryption</h4>
                <div className="form-group">
                  <Select
                    className="form-section"
                    required
                    name="encryptionType"
                    label="Encrypt using"
                    value={formData.encryptionType}
                    onChange={event => handleFormData(event, 'encryptionType')}
                  >
                    <MenuItem value="GPG">GPG</MenuItem>
                  </Select>

                  <Select
                    className="form-section"
                    required
                    name="kustomization"
                    label="kustomization"
                    value={formData.kustomization}
                    description="Choose the kustomization that will be used by SOPS to decrypt the secret."
                    onChange={event => handleFormData(event, 'kustomization')}
                  >
                    {kustomizations?.objects?.map((k, index: number) => {
                      return (
                        <MenuItem key={index} value={k.name}>
                          {k.name}
                        </MenuItem>
                      );
                    })}
                  </Select>
                </div>
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
              {/* {loading ? (
                      <LoadingPage className="create-loading" />
                    ) : (
                      <div className="create-cta">
                        <Button
                          type="submit"
                          onClick={() => setSubmitType('Create app')}
                          disabled={!enableCreatePR}
                        >
                          CREATE PULL REQUEST
                        </Button>
                      </div>
                    )}
                  </Grid> */}
              <p>{JSON.stringify(formData)}</p>
              {console.count('Form change')}
            </FormWrapper>
          )}
        </ContentWrapper>
      </CallbackStateContextProvider>
    </PageTemplate>
  );
};

export default CreateSOPS;
