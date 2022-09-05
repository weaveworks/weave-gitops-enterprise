import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { AddApplicationRequest, useApplicationsCount } from '../utils';
import GitOps from '../../Clusters/Create/Form/Partials/GitOps';
import { Grid } from '@material-ui/core';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import {
  CallbackStateContextProvider,
  clearCallbackState,
  getCallbackState,
  getProviderToken,
} from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import { isUnauthenticated, removeToken } from '../../../utils/request';
import useNotifications from '../../../contexts/Notifications';
import useProfiles from '../../../contexts/Profiles';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import { useListConfig } from '../../../hooks/versions';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import AppFields from './form/Partials/AppFields';
import Profiles from '../../Clusters/Create/Form/Partials/Profiles';
import { UpdatedProfile } from '../../../types/custom';
import ProfilesProvider from '../../../contexts/Profiles/Provider';
import { ClusterAutomation } from '../../../cluster-services/cluster_services.pb';
import useClusters from '../../../contexts/Clusters';
import { Loader } from '../../Loader';

const AddApplication = () => {
  const applicationsCount = useApplicationsCount();
  const [loading, setLoading] = useState<boolean>(false);
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const history = useHistory();
  const { setNotifications } = useNotifications();
  const { profiles } = useProfiles();
  const { data } = useListConfig();
  const repositoryURL = data?.repositoryURL || '';
  const authRedirectPage = `/applications/create`;
  const { clusters, isLoading } = useClusters();

  const random = useMemo(() => Math.random().toString(36).substring(7), []);

  let initialFormData = {
    url: '',
    provider: '',
    branchName: `add-application-branch-${random}`,
    title: 'Add application',
    commitMessage: 'Add application',
    pullRequestDescription: 'This PR adds a new application',
    clusterAutomations: [
      {
        name: '',
        namespace: '',
        cluster_name: '',
        cluster_namespace: '',
        cluster: '',
        cluster_isControlPlane: false,
        path: '',
        source_name: '',
        source_namespace: '',
        source: '',
        source_type: '',
      },
    ],
  };

  let initialProfiles = [] as UpdatedProfile[];

  const callbackState = getCallbackState();

  if (callbackState) {
    initialFormData = {
      ...initialFormData,
      ...callbackState.state.formData,
    };
    initialProfiles = [
      ...initialProfiles,
      ...callbackState.state.selectedProfiles,
    ];
  }
  const [formData, setFormData] = useState<any>(initialFormData);
  const [selectedProfiles, setSelectedProfiles] =
    useState<UpdatedProfile[]>(initialProfiles);

  useEffect(() => {
    if (repositoryURL != null) {
      setFormData((prevState: any) => ({
        ...prevState,
        url: repositoryURL,
      }));
    }
  }, [repositoryURL]);

  useEffect(() => clearCallbackState(), []);

  useEffect(() => {
    setFormData((prevState: any) => ({
      ...prevState,
      pullRequestTitle: `Add application ${(formData.clusterAutomations || [])
        .map((a: any) => a.name)
        .join(', ')}`,
    }));
  }, [formData.clusterAutomations]);

  const handleAddApplication = useCallback(() => {
    let clusterAutomations: ClusterAutomation[] = [];
    if (formData.source_type === 'HelmRepository') {
      for (let kustomization of formData.clusterAutomations) {
        for (let profile of selectedProfiles) {
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
              },
            },
          };
        },
      );
    }
    const payload = {
      head_branch: formData.branchName,
      title: formData.pullRequestTitle,
      description: formData.pullRequestDescription,
      commit_message: formData.commitMessage,
      clusterAutomations,
    };
    setLoading(true);
    return AddApplicationRequest(
      payload,
      getProviderToken(formData.provider as GitProvider),
    )
      .then(response => {
        history.push('/applications');
        setNotifications([
          {
            message: {
              component: (
                <a
                  style={{ color: weaveTheme.colors.primary }}
                  href={response.webUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  PR created successfully.
                </a>
              ),
            },
            variant: 'success',
          },
        ]);
      })
      .catch(error => {
        setNotifications([
          { message: { text: error.message }, variant: 'danger' },
        ]);
        if (isUnauthenticated(error.code)) {
          removeToken(formData.provider);
        }
      })
      .finally(() => setLoading(false));
  }, [formData, history, setNotifications, selectedProfiles]);

  return useMemo(() => {
    return (
      <ThemeProvider theme={localEEMuiTheme}>
        <PageTemplate documentTitle="WeGo · Add new application">
          <CallbackStateContextProvider
            callbackState={{
              page: authRedirectPage as PageRoute,
              state: {
                formData,
                selectedProfiles,
              },
            }}
          >
            <SectionHeader
              className="count-header"
              path={[
                {
                  label: 'Applications',
                  url: '/applications',
                  count: applicationsCount,
                },
                { label: 'Add new application' },
              ]}
            />
            <ContentWrapper>
              <Grid container>
                <Grid item xs={12} sm={10} md={10} lg={8}>
                  {!isLoading &&
                    formData.clusterAutomations.map(
                      (automation: ClusterAutomation, index: number) => {
                        return (
                          <AppFields
                            key={index}
                            index={index}
                            formData={formData}
                            setFormData={setFormData}
                            clusters={clusters}
                          />
                        );
                      },
                    )}
                  {isLoading && <Loader></Loader>}
                </Grid>
                {profiles.length > 0 &&
                formData.source_type === 'HelmRepository' ? (
                  <Profiles
                    // Temp fix to hide layers when using profiles in Add App until we update the BE
                    context="app"
                    selectedProfiles={selectedProfiles}
                    setSelectedProfiles={setSelectedProfiles}
                  />
                ) : null}
                <Grid item xs={12} sm={10} md={10} lg={8}>
                  <GitOps
                    loading={loading}
                    formData={formData}
                    setFormData={setFormData}
                    onSubmit={handleAddApplication}
                    showAuthDialog={showAuthDialog}
                    setShowAuthDialog={setShowAuthDialog}
                  />
                </Grid>
              </Grid>
            </ContentWrapper>
          </CallbackStateContextProvider>
        </PageTemplate>
      </ThemeProvider>
    );
  }, [
    applicationsCount,
    authRedirectPage,
    formData,
    handleAddApplication,
    loading,
    profiles.length,
    selectedProfiles,
    showAuthDialog,
    clusters,
    isLoading,
  ]);
};

export default () => (
  <ProfilesProvider>
    <AddApplication />
  </ProfilesProvider>
);
