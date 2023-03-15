import {
  FormControl,
  FormControlLabel,
  FormLabel,
  MenuItem,
  Radio,
  RadioGroup,
} from '@material-ui/core';
import {
  Button,
  Icon,
  IconType,
  Kind,
  Kustomization,
  theme,
} from '@weaveworks/weave-gitops';
import { useCallback, useMemo, useState } from 'react';
import styled from 'styled-components';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
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
  secretData: { key: string; value: string }[];
  secretValue: string;
  repo: string;
  provider: string;
  branchName: string;
  pullRequestTitle: string;
  commitMessage: string;
  pullRequestDescription: string;
}

const { medium } = theme.spacing;
const { neutral20, neutral10, primary10 } = theme.colors;

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
    .MuiRadio-colorSecondary.Mui-checked {
      color: ${primary10};
    }
    h2 {
      font-size: 20px;
    }
  }
  .MuiInputBase-input {
    padding-left: 8px;
  }
  .form-section {
    width: calc(40% - 24px);
    margin-right: 24px;
  }
  .auth-message {
    padding-right: 0;
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
    secretData: [],
    secretValue: '',
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
                <h2>Encryption</h2>
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
              <p>{JSON.stringify(formData)}</p>
              {console.count('Form change')}
            </FormWrapper>
          )}
        </ContentWrapper>
      </CallbackStateContextProvider>
    </PageTemplate>
  );
};

const SecretData = ({
  formData,
  handleFormData,
}: {
  formData: SOPS;
  handleFormData: (event: React.ChangeEvent<any>, key: string) => void;
}) => {
  const [type, setType] = useState('stringData');
  const [secretData, setSecretData] = useState<
    { key: string; value: string }[]
  >([{ key: '', value: '' }]);

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setType((event.target as HTMLInputElement).value);
  };
  const handleSecretChange = (index: number, isKey: boolean, value: string) => {
    const mappedData = secretData.map((e, i) => {
      if (i === index) {
        if (isKey) e.key = value;
        else e.value = value;
        return e;
      }
      return e;
    });
    setSecretData(mappedData);
  };
  return (
    <div className="group-section">
      <h2>Secret Data</h2>
      <div className="form-group">
        <FormControl>
          <RadioGroup
            row
            aria-labelledby="demo-controlled-radio-buttons-group"
            name="controlled-radio-buttons-group"
            value={type}
            onChange={handleChange}
          >
            <FormControlLabel
              value="stringData"
              control={<Radio />}
              label="String Data"
            />
            <FormControlLabel value="Data" control={<Radio />} label="Data" />
          </RadioGroup>
        </FormControl>
      </div>
      {type === 'stringData' ? (
        <Input
          className="form-section"
          required
          name="secretValue"
          label="SECRET VALUE"
          value={formData.secretValue}
          onChange={event => handleFormData(event, 'secretValue')}
        />
      ) : (
        <>
          {secretData.map((obj, index) => (
            <div key={index}>
              <Input
                className="form-section"
                required
                name="dataSecretKey"
                label="KEY"
                placeholder="secret key"
                value={obj.key}
                onChange={event =>
                  handleSecretChange(
                    index,
                    true,
                    (event.target as HTMLInputElement).value,
                  )
                }
              />
              <Input
                className="form-section"
                required
                name="dataSecretKey"
                label="VALUE"
                placeholder="secret value"
                value={obj.value}
                onChange={event =>
                  handleSecretChange(
                    index,
                    false,
                    (event.target as HTMLInputElement).value,
                  )
                }
              />
            </div>
          ))}
          <Button
            startIcon={<Icon type={IconType.AddIcon} size="base" />}
            onClick={() =>
              setSecretData(secretData => {
                return [...secretData, { key: '', value: '' }];
              })
            }
          >
            Add
          </Button>
        </>
      )}
    </div>
  );
};

const FormKeyValue = ({ handleSecretChange, index, obj }: any) => {
  return (
    <div>
      <Input
        className="form-section"
        required
        name="dataSecretKey"
        label="KEY"
        placeholder="secret key"
        value={obj.key}
        onChange={event =>
          handleSecretChange(
            index,
            true,
            (event.target as HTMLInputElement).value,
          )
        }
      />
      <Input
        className="form-section"
        required
        name="dataSecretKey"
        label="VALUE"
        placeholder="secret value"
        value={obj.value}
        onChange={event =>
          handleSecretChange(
            index,
            false,
            (event.target as HTMLInputElement).value,
          )
        }
      />
    </div>
  );
};
export default CreateSOPS;
