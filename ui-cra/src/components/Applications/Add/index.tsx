import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import styled from 'styled-components';
import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { Grid } from '@material-ui/core';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import {
  Button,
  Link,
  LoadingPage,
  useListSources,
} from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { isUnauthenticated, removeToken } from '../../../utils/request';
import useNotifications from '../../../contexts/Notifications';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import AppFields from './form/Partials/AppFields';
import {
  ClusterAutomation,
  CreateAutomationsPullRequestRequest,
  RenderAutomationResponse,
  RepositoryRef,
} from '../../../cluster-services/cluster_services.pb';
import _ from 'lodash';
import useProfiles from '../../../hooks/profiles';
import { useCallbackState } from '../../../utils/callback-state';
import { ProfilesIndex } from '../../../types/custom';
import { validateFormData } from '../../../utils/form';
import { getGitRepoHTTPSURL } from '../../../utils/formatters';
import { Routes } from '../../../utils/nav';
import Preview from '../../Templates/Form/Partials/Preview';
import Profiles from '../../Templates/Form/Partials/Profiles';
import GitOps from '../../Templates/Form/Partials/GitOps';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import {
  clearCallbackState,
  getProviderTokenHeader,
} from '../../GitAuth/utils';
import {
  ClusterFormData,
  getInitialGitRepo,
  getRepositoryUrl,
  GitopsFormData,
  SourceFormData,
} from '../../Templates/Form/utils';
import { GitRepositoryEnriched } from '../../Templates/Form';
import { getGitRepos } from '../../Clusters';
import { GitProvider } from '../../../api/gitauth/gitauth.pb';
import { EnterpriseClientContext } from '../../../contexts/EnterpriseClient';

const FormWrapper = styled.form`
  .preview-cta {
    display: flex;
    justify-content: flex-end;
    padding: ${({ theme }) => theme.spacing.small}
      ${({ theme }) => theme.spacing.base};
    button {
      width: 200px;
    }
  }
  .preview-loading {
    padding: ${({ theme }) => theme.spacing.base};
  }
  .create-cta {
    display: flex;
    justify-content: end;
    padding: ${({ theme }) => theme.spacing.small};
    button {
      width: 200px;
    }
  }
  .create-loading {
    padding: ${({ theme }) => theme.spacing.base};
  }
`;

const SourceLinkWrapper = styled.div`
  padding-top: ${({ theme }) => theme.spacing.medium};
  overflow-x: auto;
`;

function getInitialData(
  callbackState: { state: { formData: GitopsFormData } } | null,
  random: string,
): GitopsFormData {
  const defaultFormData: GitopsFormData = {
    repo: null,
    provider: '',
    branchName: `add-application-branch-${random}`,
    pullRequestTitle: 'Add application',
    commitMessage: 'Add application',
    pullRequestDescription: 'This PR adds a new application',
    source: {
      name: '',
      namespace: '',
      data: '',
      type: '',
      url: '',
      branch: '',
    },
    cluster: {
      name: '',
      namespace: '',
      data: '',
      isControlPlane: false,
    },
    clusterAutomations: [
      {
        name: '',
        namespace: '',
        target_namespace: '',
        createNamespace: false,
        path: '',
      },
    ],
  };

  const initialFormData = {
    ...defaultFormData,
    ...(callbackState?.state?.formData as GitopsFormData),
  };

  return initialFormData;
}

function toKustomization(
  cluster: ClusterFormData,
  source: SourceFormData,
  kustomization: GitopsFormData['clusterAutomations'][number],
): ClusterAutomation {
  return {
    cluster: {
      name: cluster.name,
      namespace: cluster.namespace,
    },
    isControlPlane: cluster.isControlPlane,
    kustomization: {
      metadata: {
        name: kustomization.name,
        namespace: kustomization.namespace,
      },
      spec: {
        path: kustomization.path,
        sourceRef: {
          name: source.name,
          namespace: source.namespace,
        },
        targetNamespace: kustomization.target_namespace,
        createNamespace: kustomization.createNamespace,
      },
    },
  };
}

function toHelmRelease(
  cluster: ClusterFormData,
  source: SourceFormData,
  helmRelease: GitopsFormData['clusterAutomations'][number],
  profile: ProfilesIndex[string],
  version: string,
  values: string,
): ClusterAutomation {
  return {
    cluster: {
      name: cluster.name,
      namespace: cluster.namespace,
    },
    isControlPlane: cluster.isControlPlane,
    helmRelease: {
      metadata: {
        name: profile.name,
        namespace: profile.namespace,
      },
      spec: {
        chart: {
          spec: {
            chart: profile.name,
            sourceRef: {
              name: source.name,
              namespace: source.namespace,
            },
            version,
          },
        },
        values,
      },
    },
  };
}

function getAutomations(
  cluster: ClusterFormData,
  source: SourceFormData,
  automations: GitopsFormData['clusterAutomations'],
  profiles: ProfilesIndex,
): ClusterAutomation[] {
  let clusterAutomations: ClusterAutomation[] = [];
  const selectedProfilesList = _.sortBy(Object.values(profiles), 'name').filter(
    p => p.selected,
  );

  if (source.type === 'HelmRepository') {
    for (let helmRelease of automations) {
      for (let profile of selectedProfilesList) {
        let values: string = '';
        let version: string = '';
        for (let value of profile.values) {
          if (value.selected === true) {
            version = value.version;
            values = value.yaml;
            clusterAutomations.push(
              toHelmRelease(
                cluster,
                source,
                helmRelease,
                profile,
                version,
                values,
              ),
            );
          }
        }
      }
    }
  } else {
    clusterAutomations = automations.map(ks =>
      toKustomization(cluster, source, ks),
    );
  }

  return clusterAutomations;
}

const AddApplication = ({ clusterName }: { clusterName?: string }) => {
  const [loading, setLoading] = useState<boolean>(false);
  const { api } = useContext(EnterpriseClientContext);
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const { setNotifications } = useNotifications();
  const history = useHistory();
  const authRedirectPage = `/applications/create`;
  const [formError, setFormError] = useState<string>('');

  const optionUrl = (url?: string, branch?: string) => {
    const linkText = branch ? (
      <>
        {url}@<strong>{branch}</strong>
      </>
    ) : (
      url
    );
    if (branch) {
      return (
        <Link href={getGitRepoHTTPSURL(url, branch)} newTab>
          {linkText}
        </Link>
      );
    } else {
      return (
        <Link href={getGitRepoHTTPSURL(url)} newTab>
          {linkText}
        </Link>
      );
    }
  };

  const random = useMemo(() => Math.random().toString(36).substring(7), []);

  const callbackState = useCallbackState();

  const initialFormData = getInitialData(callbackState, random);

  const [formData, setFormData] = useState<GitopsFormData>(initialFormData);
  const { cluster, source } = formData;
  const helmRepo: RepositoryRef = useMemo(() => {
    return {
      name: source.name,
      namespace: source.namespace,
      cluster: {
        name: cluster.name,
        namespace: cluster.namespace,
      },
    };
  }, [cluster, source]);

  const { profiles, isLoading: profilesIsLoading } = useProfiles(
    source.type === 'HelmRepository',
    undefined,
    undefined,
    helmRepo,
  );
  const [updatedProfiles, setUpdatedProfiles] = useState<ProfilesIndex>({});
  const [openPreview, setOpenPreview] = useState(false);
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [prPreview, setPRPreview] = useState<RenderAutomationResponse | null>(
    null,
  );
  const [enableCreatePR, setEnableCreatePR] = useState<boolean>(false);
  const { data } = useListSources();
  const gitRepos = React.useMemo(
    () => getGitRepos(data?.result),
    [data?.result],
  );
  const initialGitRepo = getInitialGitRepo(
    null,
    gitRepos,
  ) as GitRepositoryEnriched;

  useEffect(() => {
    setUpdatedProfiles({
      ..._.keyBy(profiles, 'name'),
      ...callbackState?.state?.updatedProfiles,
    });
  }, [callbackState?.state?.updatedProfiles, profiles]);

  useEffect(() => clearCallbackState(), []);

  useEffect(() => {
    setFormData(prevState => ({
      ...prevState,
      pullRequestTitle: `Add application ${(formData.clusterAutomations || [])
        .map(a => a.name)
        .join(', ')}`,
    }));
  }, [formData.clusterAutomations]);

  useEffect(() => {
    if (!formData.repo) {
      setFormData(prevState => ({
        ...prevState,
        repo: initialGitRepo,
      }));
    }
  }, [initialGitRepo, formData.repo]);

  const handlePRPreview = useCallback(() => {
    setPreviewLoading(true);
    return api
      .RenderAutomation({
        clusterAutomations: getAutomations(
          formData.cluster,
          formData.source,
          formData.clusterAutomations,
          updatedProfiles,
        ),
      })
      .then(data => {
        setOpenPreview(true);
        setPRPreview(data);
      })
      .catch(err =>
        setNotifications([
          {
            message: { text: err.message },
            severity: 'error',
            display: 'bottom',
          },
        ]),
      )
      .finally(() => setPreviewLoading(false));
  }, [
    api,
    setOpenPreview,
    setNotifications,
    formData.cluster,
    formData.source,
    formData.clusterAutomations,
    updatedProfiles,
  ]);

  const handleAddApplication = useCallback(() => {
    const payload: CreateAutomationsPullRequestRequest = {
      headBranch: formData.branchName,
      title: formData.pullRequestTitle,
      description: formData.pullRequestDescription,
      commitMessage: formData.commitMessage,
      clusterAutomations: getAutomations(
        formData.cluster,
        formData.source,
        formData.clusterAutomations,
        updatedProfiles,
      ),
      repositoryUrl: getRepositoryUrl(formData.repo!),
    };
    setLoading(true);
    return api
      .CreateAutomationsPullRequest(payload, {
        headers: getProviderTokenHeader(formData.provider as GitProvider),
      })
      .then(response => {
        setPRPreview(null);
        history.push(Routes.Applications);
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
  }, [api, formData, history, setNotifications, updatedProfiles]);

  const [submitType, setSubmitType] = useState<string>('');

  return useMemo(() => {
    return (
      <ThemeProvider theme={localEEMuiTheme}>
        <PageTemplate
          documentTitle="Add new application"
          path={[
            {
              label: 'Applications',
              url: Routes.Applications,
            },
            { label: 'Add new application' },
          ]}
        >
          <CallbackStateContextProvider
            callbackState={{
              page: authRedirectPage as PageRoute,
              state: {
                formData,
                updatedProfiles,
              },
            }}
          >
            <ContentWrapper>
              <FormWrapper
                noValidate
                onSubmit={event =>
                  validateFormData(
                    event,
                    submitType === 'PR Preview'
                      ? handlePRPreview
                      : handleAddApplication,
                    setFormError,
                    setSubmitType,
                  )
                }
              >
                <Grid container>
                  <Grid item xs={12} sm={10} md={10} lg={8}>
                    {formData.clusterAutomations.map((_, index: number) => {
                      return (
                        <AppFields
                          context="app"
                          key={index}
                          index={index}
                          formData={formData}
                          setFormData={setFormData}
                          allowSelectCluster
                          clusterName={clusterName}
                          formError={formError}
                        />
                      );
                    })}
                    {openPreview && prPreview ? (
                      <Preview
                        context="app"
                        openPreview={openPreview}
                        setOpenPreview={setOpenPreview}
                        prPreview={prPreview}
                        sourceType={formData.source.type}
                      />
                    ) : null}
                  </Grid>
                  <Grid item sm={2} md={2} lg={4}>
                    <SourceLinkWrapper>
                      {optionUrl(formData.source.url, formData.source.branch)}
                    </SourceLinkWrapper>
                  </Grid>
                  {formData.source.type === 'HelmRepository' ? (
                    <Profiles
                      cluster={{
                        name: formData.cluster.name,
                        namespace: formData.cluster.namespace,
                      }}
                      // Temp fix to hide layers when using profiles in Add App until we update the BE
                      context="app"
                      isLoading={profilesIsLoading}
                      updatedProfiles={updatedProfiles}
                      setUpdatedProfiles={setUpdatedProfiles}
                      helmRepo={helmRepo}
                    />
                  ) : null}
                  <Grid item xs={12} sm={10} md={10} lg={8}>
                    {previewLoading ? (
                      <LoadingPage className="preview-loading" />
                    ) : (
                      <div className="preview-cta">
                        <Button
                          type="submit"
                          onClick={() => setSubmitType('PR Preview')}
                        >
                          PREVIEW PR
                        </Button>
                      </div>
                    )}
                  </Grid>
                  <Grid item xs={12} sm={10} md={10} lg={8}>
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
                        <Button
                          type="submit"
                          onClick={() => setSubmitType('Create app')}
                          disabled={!enableCreatePR}
                        >
                          CREATE PULL REQUEST
                        </Button>
                      </div>
                    )}
                  </Grid>
                </Grid>
              </FormWrapper>
            </ContentWrapper>
          </CallbackStateContextProvider>
        </PageTemplate>
      </ThemeProvider>
    );
  }, [
    authRedirectPage,
    formData,
    handleAddApplication,
    loading,
    profilesIsLoading,
    updatedProfiles,
    setUpdatedProfiles,
    showAuthDialog,
    prPreview,
    openPreview,
    handlePRPreview,
    previewLoading,
    clusterName,
    enableCreatePR,
    helmRepo,
    formError,
    submitType,
  ]);
};

export default ({ ...rest }) => <AddApplication {...rest} />;
