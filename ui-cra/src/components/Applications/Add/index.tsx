import React, { useCallback, useEffect, useMemo, useState } from 'react';
import styled from 'styled-components';
import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { createDeploymentObjects, renderKustomization } from '../utils';
import { Grid } from '@material-ui/core';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import {
  Button,
  Flex,
  GitRepository,
  Link,
  LoadingPage,
  useListSources,
} from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { removeToken } from '../../../utils/request';
import useNotifications from '../../../contexts/Notifications';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import AppFields from './form/Partials/AppFields';
import {
  ClusterAutomation,
  RepositoryRef,
} from '../../../cluster-services/cluster_services.pb';
import _ from 'lodash';
import useProfiles from '../../../hooks/profiles';
import { useCallbackState } from '../../../utils/callback-state';
import {
  AppPRPreview,
  ProfilesIndex,
  ClusterPRPreview,
} from '../../../types/custom';
import { validateFormData } from '../../../utils/form';
import { getGitRepoHTTPSURL } from '../../../utils/formatters';
import { Routes } from '../../../utils/nav';
import Preview from '../../Templates/Form/Partials/Preview';
import Profiles from '../../Templates/Form/Partials/Profiles';
import GitOps from '../../Templates/Form/Partials/GitOps';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import { clearCallbackState, getProviderToken } from '../../GitAuth/utils';
import {
  useGetInitialGitRepo,
  getRepositoryUrl,
} from '../../Templates/Form/utils';
import { getGitRepos } from '../../Clusters';
import {
  expiredTokenNotification,
  useIsAuthenticated,
} from '../../../hooks/gitprovider';

const FormWrapper = styled.form`
  .preview-cta {
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

interface FormData {
  repo: GitRepository | null;
  provider: string;
  branchName: string;
  pullRequestTitle: string;
  commitMessage: string;
  pullRequestDescription: string;
  source_name: string;
  source_namespace: string;
  source: string;
  source_type: string;
  source_url: string;
  source_branch: string;
  clusterAutomations: {
    name: string;
    namespace: string;
    target_namespace: string;
    cluster_name: string;
    cluster_namespace: string;
    cluster: string;
    cluster_isControlPlane: boolean;
    createNamespace: boolean;
    path: string;
    source_name: string;
    source_namespace: string;
    source: string;
    source_type: string;
    source_url: string;
    source_branch: string;
  }[];
}

function getInitialData(
  callbackState: { state: { formData: FormData } } | null,
  random: string,
) {
  let defaultFormData = {
    repo: null,
    provider: '',
    branchName: `add-application-branch-${random}`,
    pullRequestTitle: 'Add application',
    commitMessage: 'Add application',
    pullRequestDescription: 'This PR adds a new application',
    clusterAutomations: [
      {
        name: '',
        namespace: '',
        target_namespace: '',
        cluster_name: '',
        cluster_namespace: '',
        cluster: '',
        cluster_isControlPlane: false,
        createNamespace: false,
        path: '',
        source_name: '',
        source_namespace: '',
        source: '',
        source_type: '',
        source_url: '',
        source_branch: '',
      },
    ],
  };

  const initialFormData = {
    ...defaultFormData,
    ...callbackState?.state?.formData,
  };

  return { initialFormData };
}

const AddApplication = ({ clusterName }: { clusterName?: string }) => {
  const [loading, setLoading] = useState<boolean>(false);
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

  const { initialFormData } = getInitialData(callbackState, random);

  const [formData, setFormData] = useState<any>(initialFormData);
  const firstAuto = formData.clusterAutomations[0];
  const helmRepo: RepositoryRef = useMemo(() => {
    return {
      name: firstAuto.source_name,
      namespace: firstAuto.source_namespace,
      cluster: {
        name: firstAuto.cluster_name,
        namespace: firstAuto.cluster_namespace,
      },
    };
  }, [firstAuto]);

  const { profiles, isLoading: profilesIsLoading } = useProfiles(
    firstAuto.source_type === 'HelmRepository',
    undefined,
    undefined,
    helmRepo,
  );
  const [updatedProfiles, setUpdatedProfiles] = useState<ProfilesIndex>({});
  const [openPreview, setOpenPreview] = useState(false);
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [PRPreview, setPRPreview] = useState<
    ClusterPRPreview | AppPRPreview | null
  >(null);
  const { data } = useListSources();
  const gitRepos = React.useMemo(
    () => getGitRepos(data?.result),
    [data?.result],
  );
  const initialGitRepo = useGetInitialGitRepo(null, gitRepos);

  useEffect(() => {
    setUpdatedProfiles({
      ..._.keyBy(profiles, 'name'),
      ...callbackState?.state?.updatedProfiles,
    });
  }, [callbackState?.state?.updatedProfiles, profiles]);

  useEffect(() => clearCallbackState(), []);

  useEffect(() => {
    setFormData((prevState: any) => ({
      ...prevState,
      pullRequestTitle: `Add application ${(formData.clusterAutomations || [])
        .map((a: any) => a.name)
        .join(', ')}`,
    }));
  }, [formData.clusterAutomations]);

  const getKustomizations = useCallback(() => {
    let clusterAutomations: ClusterAutomation[] = [];
    const selectedProfilesList = _.sortBy(
      Object.values(updatedProfiles),
      'name',
    ).filter(p => p.selected);
    if (formData.source_type === 'HelmRepository') {
      for (let kustomization of formData.clusterAutomations) {
        for (let profile of selectedProfilesList) {
          let values: string = '';
          let version: string = '';
          for (let value of profile.values) {
            if (value.selected === true) {
              version = value.version;
              values = value.yaml;
              clusterAutomations.push({
                cluster: {
                  name: kustomization.cluster_name,
                  namespace: kustomization.cluster_namespace,
                },
                isControlPlane: kustomization.cluster_isControlPlane,
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
                          name: formData.source_name,
                          namespace: formData.source_namespace,
                        },
                        version,
                      },
                    },
                    values,
                  },
                },
              });
            }
          }
        }
      }
    } else {
      clusterAutomations = formData.clusterAutomations.map(
        (kustomization: any) => {
          return {
            cluster: {
              name: kustomization.cluster_name,
              namespace: kustomization.cluster_namespace,
            },
            isControlPlane: kustomization.cluster_isControlPlane,
            kustomization: {
              metadata: {
                name: kustomization.name,
                namespace: kustomization.namespace,
              },
              spec: {
                path: kustomization.path,
                sourceRef: {
                  name: kustomization.source_name,
                  namespace: kustomization.source_namespace,
                },
                targetNamespace: kustomization.target_namespace,
                createNamespace: kustomization.createNamespace,
              },
            },
          };
        },
      );
    }
    return clusterAutomations;
  }, [
    formData.clusterAutomations,
    formData.source_name,
    formData.source_namespace,
    formData.source_type,
    updatedProfiles,
  ]);

  useEffect(() => {
    if (!formData.repo) {
      setFormData((prevState: any) => ({
        ...prevState,
        repo: initialGitRepo,
      }));
    }
  }, [initialGitRepo, formData.repo]);

  const handlePRPreview = useCallback(() => {
    setPreviewLoading(true);
    return renderKustomization({
      clusterAutomations: getKustomizations(),
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
  }, [setOpenPreview, getKustomizations, setNotifications]);

  const token = getProviderToken(formData.provider);

  const { isAuthenticated, validateToken } = useIsAuthenticated(
    formData.provider,
    token,
  );

  const handleAddApplication = useCallback(() => {
    const payload = {
      head_branch: formData.branchName,
      title: formData.pullRequestTitle,
      description: formData.pullRequestDescription,
      commit_message: formData.commitMessage,
      clusterAutomations: getKustomizations(),
      repositoryUrl: getRepositoryUrl(formData.repo),
      baseBranch: formData.repo.obj.spec.ref.branch
    };
    setLoading(true);
    return validateToken()
      .then(() =>
        createDeploymentObjects(payload, getProviderToken(formData.provider))
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
  }, [formData, history, getKustomizations, setNotifications, validateToken]);

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
                    {formData.clusterAutomations.map(
                      (
                        automation: FormData['clusterAutomations'][0],
                        index: number,
                      ) => {
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
                      },
                    )}
                    {openPreview && PRPreview ? (
                      <Preview
                        context="app"
                        openPreview={openPreview}
                        setOpenPreview={setOpenPreview}
                        PRPreview={PRPreview}
                        sourceType={formData.source_type}
                      />
                    ) : null}
                  </Grid>
                  <Grid item sm={2} md={2} lg={4}>
                    <SourceLinkWrapper>
                      {optionUrl(formData.source_url, formData.source_branch)}
                    </SourceLinkWrapper>
                  </Grid>
                  {formData.source_type === 'HelmRepository' ? (
                    <Profiles
                      cluster={{
                        name: formData.clusterAutomations[0].cluster_name,
                        namespace:
                          formData.clusterAutomations[0].cluster_namespace,
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
                      <Flex end className="preview-cta">
                        <Button
                          type="submit"
                          onClick={() => setSubmitType('PR Preview')}
                        >
                          PREVIEW PR
                        </Button>
                      </Flex>
                    )}
                  </Grid>
                  <Grid item xs={12} sm={10} md={10} lg={8}>
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
                        <Button
                          type="submit"
                          onClick={() => setSubmitType('Create app')}
                          disabled={!isAuthenticated}
                        >
                          CREATE PULL REQUEST
                        </Button>
                      </Flex>
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
    PRPreview,
    openPreview,
    handlePRPreview,
    previewLoading,
    clusterName,
    helmRepo,
    formError,
    submitType,
    isAuthenticated,
  ]);
};

export default ({ ...rest }) => <AddApplication {...rest} />;
