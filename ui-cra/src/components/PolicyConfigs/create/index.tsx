import { MenuItem } from '@material-ui/core';
import { ThemeProvider } from '@material-ui/core/styles';
import {
  Button,
  GitRepository,
  Link,
  LoadingPage,
  theme,
  useListSources,
} from '@weaveworks/weave-gitops';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import {
  ClusterAutomation,
  GitopsCluster,
  PolicyConfigMatch,
  PolicyConfigPolicy,
} from '../../../cluster-services/cluster_services.pb';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import useNotifications from '../../../contexts/Notifications';
import { localEEMuiTheme } from '../../../muiTheme';
import { useCallbackState } from '../../../utils/callback-state';
import { Input, Select, validateFormData } from '../../../utils/form';
import { Routes } from '../../../utils/nav';
import { isUnauthenticated, removeToken } from '../../../utils/request';
import {
  CreateDeploymentObjects,
  useClustersWithSources,
} from '../../Applications/utils';
import { getGitRepos } from '../../Clusters';
import { clearCallbackState, getProviderToken } from '../../GitAuth/utils';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SelectSecretStore } from '../../Secrets/Create/Form/Partials/SelectSecretStore';
import { PreviewPRModal } from '../../Secrets/Create/PreviewPRModal';
import { GitRepositoryEnriched } from '../../Templates/Form';
import GitOps from '../../Templates/Form/Partials/GitOps';
import {
  getInitialGitRepo,
  getRepositoryUrl,
} from '../../Templates/Form/utils';

const { medium, large } = theme.spacing;
const { neutral20, neutral10 } = theme.colors;

const FormWrapper = styled.form`
  width: 80%;
  padding-bottom: ${large} !important;
  .group-section {
    border-bottom: 1px dotted ${neutral20};
    .form-section {
      width: 100%;

      .Mui-disabled {
        background: ${neutral10} !important;
        border-color: ${neutral20} !important;
      }
      .MuiInputBase-root {
        width: 50%;
      }
      .MuiFormControl-root {
        width: 50%;
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
    policyConfigName: string;
    matchType: string;
    match?: PolicyConfigMatch;
    policies: PolicyConfigPolicy[];
    isControlPlane: boolean;
    clusterName: string;
    clusterNamespace: string;
    selectedCluster: any;
  }[];
}

function getInitialData(
  callbackState: { state: { formData: FormData } } | null,
  random: string,
) {
  let defaultFormData = {
    repo: null,
    provider: '',
    branchName: `add-policyConfig-branch-${random}`,
    pullRequestTitle: 'Add PolicyConfig',
    commitMessage: 'Add PolicyConfig',
    pullRequestDescription: 'This PR adds a new PolicyConfig',
    clusterAutomations: [
      {
        clusterName: '',
        clusterNamespace: '',
        isControlPlane: false,
        policyConfigName: '',
        secretNamespace: '',
        matchType: '',
      },
    ],
  };

  const initialFormData = {
    ...defaultFormData,
    ...callbackState?.state?.formData,
  };

  return { initialFormData };
}
const CreatePolicyConfig = () => {
  const history = useHistory();

  let clusters: GitopsCluster[] | undefined = useClustersWithSources(true);
  const { setNotifications } = useNotifications();

  const callbackState = useCallbackState();
  const random = useMemo(() => Math.random().toString(36).substring(7), []);
  const { initialFormData } = getInitialData(callbackState, random);
  const authRedirectPage = `/policyConfigs/create`;

  const [loading, setLoading] = useState<boolean>(false);
  const [isclusterSelected, setIsclusterSelected] = useState<boolean>(false);

  const [showAuthDialog, setShowAuthDialog] = useState<boolean>(false);
  const [formData, setFormData] = useState<any>(initialFormData);
  const [policiesList, setPoliciesList] = useState<PolicyConfigPolicy[]>();
  const [selectedPolicies, setSelectedPolicies] =
    useState<PolicyConfigPolicy[]>();
  const [enableCreatePR, setEnableCreatePR] = useState<boolean>(false);

  const { data } = useListSources();
  const gitRepos = useMemo(() => getGitRepos(data?.result), [data?.result]);
  const initialGitRepo = getInitialGitRepo(
    null,
    gitRepos,
  ) as GitRepositoryEnriched;

  const [formError, setFormError] = useState<string>('');
  const automation = formData.clusterAutomations[0];

  const {
    clusterName,
    policyConfigName,
    clusterNamespace,
    isControlPlane,
    matchType,
    match,
  } = automation;

  useEffect(() => clearCallbackState(), []);

  useEffect(() => {
    if (!formData.repo) {
      setFormData((prevState: any) => ({
        ...prevState,
        repo: initialGitRepo,
      }));
    }
    if (clusterName) {
      setIsclusterSelected(true);
    }
  }, [initialGitRepo, formData.repo, clusterName]);

  const HandleSelectCluster = (event: React.ChangeEvent<any>) => {
    const cluster = event.target.value;
    const value = JSON.parse(cluster);
    let currentAutomation = [...formData.clusterAutomations];
    setSelectedPolicies([]);
    currentAutomation[0] = {
      ...automation,
      isControlPlane: value.name === 'management' ? true : false,
      clusterName: value,
      clusterNamespace: value.namespace,
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
      pullRequestTitle: `Add PolicyConfig ${formData.clusterAutomations[0].policyConfigName}`,
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
      policyConfig: {
        metadata: {
          name: policyConfigName,
        },
        spec: {
          match: {
            [matchType]: match[matchType],
          },
          config: {},
        },
      },
    });
    return clusterAutomations;
  }, [
    clusterName,
    clusterNamespace,
    isControlPlane,
    policyConfigName,
    matchType,
    match,
  ]);

  const handleCreatePolicyConfig = useCallback(() => {
    const payload = {
      headBranch: formData.branchName,
      title: formData.pullRequestTitle,
      description: formData.pullRequestDescription,
      commitMessage: formData.commitMessage,
      clusterAutomations: getClusterAutomations(),
      repositoryUrl: getRepositoryUrl(formData.repo),
    };
    console.log(payload);
    setLoading(true);
    setLoading(false);

    return CreateDeploymentObjects(
      payload,
      getProviderToken(formData.provider),
    );
    //   .then(response => {
    //     history.push(Routes.Secrets);
    //     setNotifications([
    //       {
    //         message: {
    //           component: (
    //             <Link href={response.webUrl} newTab>
    //               PR created successfully, please review and merge the pull
    //               request to apply the changes to the cluster.
    //             </Link>
    //           ),
    //         },
    //         severity: 'success',
    //       },
    //     ]);
    //   })
    //   .catch(error => {
    //     setNotifications([
    //       {
    //         message: { text: error.message },
    //         severity: 'error',
    //         display: 'bottom',
    //       },
    //     ]);
    //     if (isUnauthenticated(error.code)) {
    //       removeToken(formData.provider);
    //     }
    //   })
    //   .finally(() => setLoading(false));
  }, [formData, getClusterAutomations, history, setNotifications]);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate
        documentTitle="Secrets"
        path={[
          { label: 'PolicyConfigs', url: Routes.PolicyConfigs },
          { label: 'Create new PolicyConfig' },
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
          <ContentWrapper>
            <FormWrapper
              noValidate
              onSubmit={event =>
                validateFormData(event, handleCreatePolicyConfig, setFormError)
              }
            >
              <div className="group-section">
                <div className="form-group">
                  <Input
                    className="form-section"
                    required
                    name="policyConfigName"
                    label="NAME"
                    value={policyConfigName}
                    onChange={event =>
                      handleFormData(event, 'policyConfigName')
                    }
                    error={
                      formError === 'policyConfigName' && !policyConfigName
                    }
                  />
                  <Select
                    className="form-section"
                    name="clusterName"
                    required={true}
                    label="CLUSTER"
                    value={clusterName || ''}
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
                {/* {isclusterSelected && (
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
                /> */}
                {/* <Input
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
                /> */}
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
    </ThemeProvider>
  );
};

export default CreatePolicyConfig;
