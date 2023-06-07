import { MenuItem } from '@material-ui/core';
import { ThemeProvider } from '@material-ui/core/styles';
import {
  Button,
  GitRepository,
  Link,
  LoadingPage,
  Page,
  useListSources,
} from '@weaveworks/weave-gitops';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import {
  ClusterAutomation,
  ExternalSecretStore,
  GitopsCluster,
} from '../../../cluster-services/cluster_services.pb';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import useNotifications from '../../../contexts/Notifications';
import {
  expiredTokenNotification,
  useIsAuthenticated,
} from '../../../hooks/gitprovider';
import { localEEMuiTheme } from '../../../muiTheme';
import { useCallbackState } from '../../../utils/callback-state';
import { Input, Select, validateFormData } from '../../../utils/form';
import { Routes } from '../../../utils/nav';
import { removeToken } from '../../../utils/request';
import {
  createDeploymentObjects,
  useClustersWithSources,
} from '../../Applications/utils';
import { getGitRepos } from '../../Clusters';
import { clearCallbackState, getProviderToken } from '../../GitAuth/utils';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import GitOps from '../../Templates/Form/Partials/GitOps';
import {
  getRepositoryUrl,
  useGetInitialGitRepo,
} from '../../Templates/Form/utils';
import { SelectSecretStore } from './Form/Partials/SelectSecretStore';
import { PreviewPRModal } from './PreviewPRModal';

const FormWrapper = styled.form`
  width: 80%;
  padding-bottom: ${props => props.theme.spacing.large} !important;
  .group-section {
    border-bottom: 1px dotted ${props => props.theme.colors.neutral20};
    .form-section {
      width: 50%;
      .Mui-disabled {
        background: ${props => props.theme.colors.neutral10} !important;
        border-color: ${props => props.theme.colors.neutral20} !important;
      }
      .MuiInputBase-root {
        margin-right: ${props => props.theme.spacing.medium};
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
    targetCluster: string;
    clusterName: string;
    clusterNamespace: string;
    isControlPlane: boolean;
    secretName: string;
    secretNamespace: string;
    secretStoreKind: string;
    secretStoreRef: string;
    dataSecretKey: string;
    dataRemoteRefKey: string;
    dataRemoteRef_property: string;
  }[];
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
        clusterName: '',
        clusterNamespace: '',
        isControlPlane: false,
        secretName: '',
        secretNamespace: '',
        refreshInterval: '1h',
        secretStoreRef: '',
        secretStoreType: '',
        dataSecretKey: '',
        dataRemoteRefKey: '',
        dataRemoteRef_property: '',
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

  let clusters: GitopsCluster[] | undefined = useClustersWithSources(true);
  const { setNotifications } = useNotifications();

  const callbackState = useCallbackState();
  const random = useMemo(() => Math.random().toString(36).substring(7), []);
  const { initialFormData } = getInitialData(callbackState, random);
  const authRedirectPage = `/secrets/create`;

  const [loading, setLoading] = useState<boolean>(false);
  const [isclusterSelected, setIsclusterSelected] = useState<boolean>(false);

  const [showAuthDialog, setShowAuthDialog] = useState<boolean>(false);
  const [formData, setFormData] = useState<any>(initialFormData);
  const [selectedSecretStore, setSelectedSecretStore] =
    useState<ExternalSecretStore>({});

  const { data } = useListSources();
  const gitRepos = useMemo(() => getGitRepos(data?.result), [data?.result]);
  const initialGitRepo = useGetInitialGitRepo(null, gitRepos);

  const [formError, setFormError] = useState<string>('');
  const automation = formData.clusterAutomations[0];

  const {
    secretName,
    secretNamespace,
    secretStoreKind,
    dataSecretKey,
    clusterName,
    secretStoreRef,
    dataRemoteRefKey,
    dataRemoteRef_property,
    clusterNamespace,
    isControlPlane,
    targetCluster,
  } = automation;

  useEffect(() => clearCallbackState(), []);

  useEffect(() => {
    if (!formData.repo) {
      setFormData((prevState: any) => ({
        ...prevState,
        repo: initialGitRepo,
      }));
    }
    if (targetCluster) {
      setIsclusterSelected(true);
    }
  }, [initialGitRepo, formData.repo, targetCluster]);

  const HandleSelectCluster = (event: React.ChangeEvent<any>) => {
    const cluster = event.target.value;
    const value = JSON.parse(cluster);
    let currentAutomation = [...formData.clusterAutomations];
    setSelectedSecretStore({});
    currentAutomation[0] = {
      ...automation,
      isControlPlane: value.name === 'management' ? true : false,
      clusterName: value.name,
      clusterNamespace: value.namespace,
      targetCluster: cluster,
      secretStoreType: '',
      secretStoreRef: '',
      secretNamespace: '',
      secretStoreKind: '',
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
      pullRequestTitle: `Add External Secret ${formData.clusterAutomations[0].secretName}`,
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

  const getClusterAutomations = useCallback(() => {
    let clusterAutomations: ClusterAutomation[] = [];
    clusterAutomations.push({
      cluster: {
        name: clusterName,
        namespace: clusterNamespace,
      },
      isControlPlane: isControlPlane,
      externalSecret: {
        metadata: {
          name: secretName,
          namespace: secretNamespace,
        },
        spec: {
          refreshInterval: '1h',
          secretStoreRef: {
            name: secretStoreRef,
            kind: secretStoreKind,
          },
          target: {
            name: dataSecretKey,
          },
          data: {
            secretKey: dataSecretKey,
            remoteRef: {
              key: dataRemoteRefKey,
              property: dataRemoteRef_property,
            },
          },
        },
      },
    });
    return clusterAutomations;
  }, [
    clusterName,
    clusterNamespace,
    isControlPlane,
    secretName,
    secretNamespace,
    secretStoreRef,
    secretStoreKind,
    dataSecretKey,
    dataRemoteRefKey,
    dataRemoteRef_property,
  ]);

  const token = getProviderToken(formData.provider);

  const { isAuthenticated, validateToken } = useIsAuthenticated(
    formData.provider,
    token,
  );

  const handleCreateSecret = useCallback(() => {
    const payload = {
      headBranch: formData.branchName,
      title: formData.pullRequestTitle,
      description: formData.pullRequestDescription,
      commitMessage: formData.commitMessage,
      clusterAutomations: getClusterAutomations(),
      repositoryUrl: getRepositoryUrl(formData.repo),
      baseBranch: formData.repo.obj.spec.ref.branch
    };
    setLoading(true);
    return validateToken()
      .then(() =>
        createDeploymentObjects(payload, getProviderToken(formData.provider))
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
          .catch(error =>
            setNotifications([
              {
                message: { text: error.message },
                severity: 'error',
                display: 'bottom',
              },
            ]),
          )
          .finally(() => setLoading(false)),
      )
      .catch(() => {
        removeToken(formData.provider);
        setNotifications([expiredTokenNotification]);
      })
      .finally(() => setLoading(false));
  }, [
    formData,
    getClusterAutomations,
    history,
    setNotifications,
    validateToken,
  ]);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <Page
        path={[
          { label: 'Secrets', url: Routes.Secrets },
          { label: 'Create new external secret' },
        ]}
      >
        <CallbackStateContextProvider
          callbackState={{
            page: authRedirectPage as PageRoute,
            state: {
              formData,
            },
          }}
        >
          <NotificationsWrapper>
            <FormWrapper
              noValidate
              onSubmit={event =>
                validateFormData(event, handleCreateSecret, setFormError)
              }
            >
              <div className="group-section">
                <div className="form-group">
                  <Input
                    className="form-section"
                    required
                    name="secretName"
                    label="EXTERNAL SECRET NAME"
                    value={secretName}
                    onChange={event => handleFormData(event, 'secretName')}
                    error={formError === 'secretName' && !secretName}
                  />
                  <Input
                    className="form-section"
                    required
                    name="dataSecretKey"
                    label="TARGET K8s SECRET NAME"
                    value={dataSecretKey}
                    onChange={event => handleFormData(event, 'dataSecretKey')}
                    error={formError === 'dataSecretKey' && !dataSecretKey}
                  />
                  <Select
                    className="form-section"
                    name="clusterName"
                    required={true}
                    label="TARGET CLUSTER"
                    value={targetCluster || ''}
                    onChange={HandleSelectCluster}
                    error={formError === 'clusterName' && !clusterName}
                  >
                    {!clusters?.length ? (
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
                    cluster={
                      clusterNamespace
                        ? `${clusterNamespace}/${clusterName}`
                        : clusterName
                    }
                    formError={formError}
                    handleFormData={handleFormData}
                    selectedSecretStore={selectedSecretStore || {}}
                    setSelectedSecretStore={setSelectedSecretStore}
                    formData={formData}
                    setFormData={setFormData}
                    automation={automation}
                  />
                )}
                <Input
                  className="form-section"
                  required
                  name="dataRemoteRefKey"
                  label="SECRET PATH"
                  value={dataRemoteRefKey}
                  onChange={event => handleFormData(event, 'dataRemoteRefKey')}
                  error={formError === 'dataRemoteRefKey' && !dataRemoteRefKey}
                />
                <Input
                  className="form-section"
                  required
                  name="dataRemoteRef_property"
                  label="PROPERTY"
                  value={dataRemoteRef_property}
                  onChange={event =>
                    handleFormData(event, 'dataRemoteRef_property')
                  }
                  error={
                    formError === 'dataRemoteRef_property' &&
                    !dataRemoteRef_property
                  }
                />
              </div>
              <PreviewPRModal
                formData={formData}
                getClusterAutomations={getClusterAutomations}
              />
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
                <div className="create-cta">
                  <Button type="submit" disabled={!isAuthenticated}>
                    CREATE PULL REQUEST
                  </Button>
                </div>
              )}
            </FormWrapper>
          </NotificationsWrapper>
        </CallbackStateContextProvider>
      </Page>
    </ThemeProvider>
  );
};

export default CreateSecret;
