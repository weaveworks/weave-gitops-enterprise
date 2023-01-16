import { MenuItem } from '@material-ui/core';
import { useCallback, useEffect, useMemo, useState } from 'react';
import styled from 'styled-components';
import { ClusterAutomation } from '../../../cluster-services/cluster_services.pb';
import useClusters from '../../../hooks/clusters';
import { useCallbackState } from '../../../utils/callback-state';
import { Input, Select, validateFormData } from '../../../utils/form';
import useNotifications from '../../../contexts/Notifications';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import GitOps from '../../Templates/Form/Partials/GitOps';
import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { useGetSecretStoreDetails } from '../../../contexts/Secrets';
import { useListConfigContext } from '../../../contexts/ListConfig';
import { useHistory } from 'react-router-dom';
import {
  Button,
  getProviderToken,
  Link,
  LoadingPage,
  theme,
  GitRepository,
  useListSources,

} from '@weaveworks/weave-gitops';
import { isUnauthenticated, removeToken } from '../../../utils/request';
import { AddApplicationRequest } from '../../Applications/utils';
import {
  getInitialGitRepo,
  getRepositoryUrl,
} from '../../Templates/Form/utils';
import { GitRepositoryEnriched } from '../../Templates/Form';
import { getGitRepos } from '../../Clusters';

const { medium, large } = theme.spacing;
const { neutral20, neutral10 } = theme.colors;

const FormWrapper = styled.form`
  width: 80%;
  padding-bottom: ${large} !important;
  .group-section {
    border-bottom: 1px dotted ${neutral20};
    .form-section {
      width: 50%;
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

interface FormData {
  repo: GitRepository | null;
  branchName: string;
  provider: string;
  pullRequestTitle: string;
  commitMessage: string;
  pullRequestDescription: string;
  clusterAutomations: {
    target_Cluster: string;
    cluster_name: string;
    cluster_namespace: string;
    isControlPlane: boolean;
    secret_name: string;
    secret_namespace: string;
    secretStoreValue: string;
    secretStoreRef: string;
    data_secretKey: string;
    data_remoteRef_key: string;
    data_remoteRef_property: string;
  }[];
}

interface SelectSecretStoreProps {
  cluster: string;
  secretStoreValue: string;
  handleSelectSecretStore: (event: React.ChangeEvent<any>) => void;
  formError: string;
  secretStoreRef: string;
  secretStoreKind: string;
  secret_namespace: string;
}

function getInitialData(
  callbackState: { state: { formData: FormData } } | null,
  random: string,
) {
  let defaultFormData = {
    repo: null,
    provider: '',
    branchName: `add-external-secret-branch-${random}`,
    pullRequestTitle: 'Add External Secret',
    commitMessage: 'Add External Secret',
    pullRequestDescription: 'This PR adds a new External Secret',
    clusterAutomations: [
      {
        cluster_name: '',
        cluster_namespace: '',
        isControlPlane: false,
        secret_name: '',
        secret_namespace: '',
        refreshInterval: '1h',
        secretStoreRef: '',
        secretStoreKind: '',
        data_secretKey: '',
        data_remoteRef_key: '',
        data_remoteRef_property: '',
      },
    ],
  };

  const initialFormData = {
    ...defaultFormData,
    ...callbackState?.state?.formData,
  };

  return { initialFormData };
}
const CreateSecret = () => {
  const history = useHistory();

  const { clusters, isLoading: isClustersLoading } = useClusters();
  const { setNotifications } = useNotifications();

  const callbackState = useCallbackState();
  const random = useMemo(() => Math.random().toString(36).substring(7), []);
  const { initialFormData } = getInitialData(callbackState, random);


  const [loading, setLoading] = useState<boolean>(false);
  const [isclusterSelected, setIsclusterSelected] = useState<boolean>(false);

  const [showAuthDialog, setShowAuthDialog] = useState<boolean>(false);
  const [formData, setFormData] = useState<any>(initialFormData);
  const [submitType, setSubmitType] = useState<string>('');

  const [enableCreatePR, setEnableCreatePR] = useState<boolean>(false);

  const { data } = useListSources();
  const gitRepos = useMemo(
    () => getGitRepos(data?.result),
    [data?.result],
  );
  const initialGitRepo = getInitialGitRepo(
    null,
    gitRepos,
  ) as GitRepositoryEnriched;

  const [formError, setFormError] = useState<string>('');
  const automation = formData.clusterAutomations[0];

  const {
    secret_name,
    secret_namespace,
    data_secretKey,
    cluster_name,
    secretStoreRef,
    secretStoreKind,
    data_remoteRef_key,
    data_remoteRef_property,
    cluster_namespace,
    isControlPlane,
    secretStoreValue,
    target_Cluster,
  } = automation;

  useEffect(() => {
    if (!formData.repo) {
      setFormData((prevState: any) => ({
        ...prevState,
        repo: initialGitRepo,
      }));
    }
  }, [initialGitRepo, formData.repo]);


  const handleSelectSecretStore = (event: React.ChangeEvent<any>) => {
    const sercetStore = event.target.value;
    const value = JSON.parse(sercetStore);
    let currentAutomation = [...formData.clusterAutomations];
    currentAutomation[0] = {
      ...automation,
      secretStoreRef: value.name,
      secret_namespace: value.namespace,
      secretStoreKind: value.kind,
      secretStoreValue: value,
    };
    setFormData({
      ...formData,
      clusterAutomations: currentAutomation,
    });
  };

  const HandleSelectCluster = (event: React.ChangeEvent<any>) => {
    const cluster = event.target.value;
    const value = JSON.parse(cluster);

    let currentAutomation = [...formData.clusterAutomations];
    currentAutomation[0] = {
      ...automation,
      isControlPlane: value.name === 'management' ? true : false,
      cluster_name: value.name,
      cluster_namespace: value.namespace,
      target_Cluster: value,
    };
    setFormData({
      ...formData,
      clusterAutomations: currentAutomation,
    });
    setIsclusterSelected(true);
  };

  useEffect(() => {
    setFormData((prevState: any) => ({
      ...prevState,
      pullRequestTitle: `Add External Secret ${formData.clusterAutomations[0].secret_name}`,
    }));
  }, [formData.clusterAutomations]);

  const handleFormData = (
    event: React.ChangeEvent<{ name?: string; value: unknown }>,
    fieldName?: string,
  ) => {
    const { value } = event?.target;
    let currentAutomation = [...formData.clusterAutomations];

    currentAutomation[0] = {
      ...automation,
      [fieldName as string]: value,
    };

    setFormData({
      ...formData,
      clusterAutomations: currentAutomation,
    });
  };


  const handleCreateSecret = useCallback(() => {
    let clusterAutomations: ClusterAutomation[] = [];
    clusterAutomations.push({
      cluster: {
        name: cluster_name,
        namespace: cluster_namespace,
      },
      isControlPlane: isControlPlane,
      externalSecret: {
        metadata: {
          name: secret_name,
          namespace: secret_namespace,
        },
        spec: {
          refreshInterval: '1h',
          secretStoreRef: {
            name: secretStoreRef,
          },
          target: {
            name: data_secretKey,
          },
          data: {
            secretKey: data_secretKey,
            remoteRef: {
              key: data_remoteRef_key,
              property: data_remoteRef_property,
            },
          },
        },
      },
    });
    const payload = {
      headBranch: formData.branchName,
      title: formData.pullRequestTitle,
      description: formData.pullRequestDescription,
      commitMessage: formData.commitMessage,
      clusterAutomations: clusterAutomations,
      repositoryUrl: getRepositoryUrl(formData.repo),
    };
    setLoading(true);
    return AddApplicationRequest(payload, getProviderToken(formData.provider))
      .then(response => {
        history.push(Routes.Secrets);
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
      })
      .catch(error => {
        setNotifications([
          {
            message: { text: error.message },
            severity: 'error',
            display: 'bottom',
          },
        ]);
        if (isUnauthenticated(error.code)) {
          removeToken(formData.provider);
        }
      })
      .finally(() => setLoading(false));
  }, [formData]);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate
        documentTitle="Secrets"
        path={[
          { label: 'Secrets', url: Routes.Secrets },
          { label: 'Create new secret' },
        ]}
      >
        <ContentWrapper>
          <FormWrapper
            noValidate
            onSubmit={event =>
              validateFormData(
                event,
                handleCreateSecret,
                setFormError,
                setSubmitType,
              )
            }
          >
            <div className="group-section">
              <div className="form-group">
                <Input
                  className="form-section"
                  required
                  name="secret_name"
                  label="EXTERNAL SECRET NAME"
                  value={secret_name}
                  onChange={event => handleFormData(event, 'secret_name')}
                  error={formError === 'secret_name' && !secret_name}
                />
                <Input
                  className="form-section"
                  required
                  name="data_secretKey"
                  label="TARGET K8s SECRET NAME"
                  value={data_secretKey}
                  onChange={event => handleFormData(event, 'data_secretKey')}
                  error={formError === 'data_secretKey' && !data_secretKey}
                />
                <Select
                  className="form-section"
                  name="cluster_name"
                  required={true}
                  label="TARGET CLUSTER"
                  value={target_Cluster}
                  onChange={HandleSelectCluster}
                  error={formError === 'cluster_name' && !cluster_name}
                >
                  {isClustersLoading ? (
                    <MenuItem disabled={true}>Loading...</MenuItem>
                  ) : (
                    clusters?.map((option, index: number) => {
                      return (
                        <MenuItem key={index} value={JSON.stringify(option)}>
                          {option.name}
                        </MenuItem>
                      );
                    })
                  )}
                </Select>
              </div>
              {isclusterSelected && (
                <SelectSecretStore
                  cluster={cluster_name}
                  secretStoreValue={secretStoreValue}
                  handleSelectSecretStore={handleSelectSecretStore}
                  formError={formError}
                  secretStoreRef={secretStoreRef}
                  secretStoreKind={secretStoreKind}
                  secret_namespace={secret_namespace}
                />
              )}
              <Input
                className="form-section"
                required
                name="data_remoteRef_key"
                label="SECRET PATH"
                value={data_remoteRef_key}
                onChange={event => handleFormData(event, 'data_remoteRef_key')}
                error={
                  formError === 'data_remoteRef_key' && !data_remoteRef_key
                }
              />
              <Input
                className="form-section"
                required
                name="data_remoteRef_property"
                label="PROPERTY"
                value={data_remoteRef_property}
                onChange={event =>
                  handleFormData(event, 'data_remoteRef_property')
                }
                error={
                  formError === 'data_remoteRef_property' &&
                  !data_remoteRef_property
                }
              />
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
      </PageTemplate>
    </ThemeProvider>
  );
};

export default CreateSecret;

export const SelectSecretStore = (props: SelectSecretStoreProps) => {
  const {
    cluster,
    secretStoreValue,
    handleSelectSecretStore,
    formError,
    secretStoreRef,
    secretStoreKind,
    secret_namespace,
  } = props;
  const { data, isLoading } = useGetSecretStoreDetails({
    clusterName: cluster,
  });
  const secretStores = data?.stores;
  return (
    <div className="form-group">
      <Select
        className="form-section"
        name="secretStoreRef"
        required
        label="SECRET STORE"
        value={secretStoreValue}
        onChange={handleSelectSecretStore}
        error={formError === 'secretStoreRef' && !secretStoreRef}
      >
        {isLoading ? (
          <MenuItem disabled={true}>Loading...</MenuItem>
        ) : secretStores?.length ? (
          secretStores.map((option, index: number) => {
            return (
              <MenuItem key={index} value={JSON.stringify(option)}>
                {option.name}
              </MenuItem>
            );
          })
        ) : (
          <MenuItem disabled={true}>
            No SecretStore found on that cluster
          </MenuItem>
        )}
      </Select>
      <Input
        className="form-section"
        name="secret_store_kind"
        label="SECRET STORE KIND"
        value={secretStoreKind}
        disabled={true}
        error={false}
      />
      <Input
        className="form-section"
        required
        name="secret_namespace"
        label="TARGET NAMESPACE"
        value={secret_namespace}
        disabled
        error={false}
      />
    </div>
  );
};
