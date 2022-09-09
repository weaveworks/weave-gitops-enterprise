import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { AddApplicationRequest, useApplicationsCount } from '../utils';
import GitOps from '../../Clusters/Form/Partials/GitOps';
import { Grid } from '@material-ui/core';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import {
  CallbackStateContextProvider,
  getProviderToken,
  GitRepository,
  HelmRepository,
} from '@weaveworks/weave-gitops';
import { useHistory, useParams } from 'react-router-dom';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import { isUnauthenticated, removeToken } from '../../../utils/request';
import useNotifications from '../../../contexts/Notifications';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import { useListConfig } from '../../../hooks/versions';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import AppFields from './form/Partials/AppFields';
import Profiles from '../../Clusters/Form/Partials/Profiles';
import ProfilesProvider from '../../../contexts/Profiles/Provider';
import {
  ClusterAutomation,
  CreateAutomationsPullRequestRequest,
  GitopsCluster,
} from '../../../cluster-services/cluster_services.pb';
import _ from 'lodash';
import useProfiles from '../../../contexts/Profiles';
import { useCallbackState } from '../../../utils/callback-state';
import { ProfilesIndex } from '../../../types/custom';
import { Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import { useSearchParam } from 'react-use';

export interface ClusterAutomationFormData {
  name: string;
  namespace: string;
  clusterName: string | null;
  path: string;
  source: HelmRepository | GitRepository | null;
}

export interface AddAppFormData {
  url: string;
  provider: string;
  branchName: string;
  title: string;
  commitMessage: string;
  pullRequestTitle: string;
  pullRequestDescription: string;
  clusterAutomations: ClusterAutomationFormData[];
}

const toCluster = (clusterName: string): GitopsCluster => {
  const [firstBit, secondBit] = clusterName.split('/');
  const [namespace, name, controlPlane] = secondBit
    ? [firstBit, secondBit, false]
    : ['', firstBit, true];
  return {
    name,
    namespace,
    controlPlane,
  };
};

const toClusterName = (cluster: GitopsCluster): string => {
  return cluster.namespace
    ? `${cluster.namespace}/${cluster.name}`
    : `${cluster.name}`;
};

const toPayload = (
  formData: AddAppFormData,
  updatedProfiles: ProfilesIndex,
): CreateAutomationsPullRequestRequest => {
  const automation = formData.clusterAutomations?.[0];
  const cluster = toCluster(automation.clusterName!);

  let clusterAutomations: ClusterAutomation[] = [];
  if (automation?.source?.kind === 'KindHelmRepository') {
    const selectedProfilesList = _.sortBy(
      Object.values(updatedProfiles),
      'name',
    ).filter(p => p.selected);

    for (let profile of selectedProfilesList) {
      for (let value of profile.values) {
        if (value.selected === true) {
          const version = value.version;
          const values = value.yaml;
          clusterAutomations.push({
            cluster: {
              name: cluster.name,
              namespace: cluster.namespace,
            },
            isControlPlane: cluster.controlPlane,
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
                      name: automation.source.name,
                      namespace: automation.source.namespace,
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
  return {
    headBranch: formData.branchName,
    title: formData.pullRequestTitle,
    description: formData.pullRequestDescription,
    commitMessage: formData.commitMessage,
    clusterAutomations,
  };
};

const AddApplication = () => {
  const applicationsCount = useApplicationsCount();
  const [loading, setLoading] = useState<boolean>(false);
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const history = useHistory();
  const { setNotifications } = useNotifications();
  const { data } = useListConfig();
  const repositoryURL = data?.repositoryURL || '';
  const authRedirectPage = `/applications/create`;

  const random = useMemo(() => Math.random().toString(36).substring(7), []);

  const callbackState = useCallbackState();

  const initialFormData: AddAppFormData = {
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
        path: '',
        clusterName: useSearchParam('clusterName'),
        source: null,
      },
    ],
    ...callbackState?.state?.formData,
  };

  const [formData, setFormData] = useState<AddAppFormData>(initialFormData);
  const automation = formData.clusterAutomations?.[0];
  console.log({ automation });

  const { profiles, isLoading: profilesIsLoading } = useProfiles();
  const [updatedProfiles, setUpdatedProfiles] = useState<ProfilesIndex>({});

  useEffect(() => {
    if (automation?.clusterName) {
      const params = new URLSearchParams();
      params.set('clusterName', automation?.clusterName!);
      history.push({ search: params.toString() });
    }
  }, [automation?.clusterName, history]);

  useEffect(() => {
    setUpdatedProfiles({
      ..._.keyBy(profiles, 'name'),
      ...callbackState?.state?.updatedProfiles,
    });
  }, [callbackState?.state?.updatedProfiles, profiles]);

  useEffect(() => {
    setFormData((prevState: any) => ({
      ...prevState,
      url: repositoryURL,
    }));
  }, [repositoryURL]);

  useEffect(() => {
    setFormData((prevState: any) => ({
      ...prevState,
      pullRequestTitle: `Add application ${(formData.clusterAutomations || [])
        .map((a: any) => a.name)
        .join(', ')}`,
    }));
  }, [formData.clusterAutomations]);

  const handleAddApplication = useCallback(() => {
    const payload = toPayload(formData, updatedProfiles);
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
  }, [formData, history, setNotifications, updatedProfiles]);

  return useMemo(() => {
    return (
      <ThemeProvider theme={localEEMuiTheme}>
        <PageTemplate documentTitle="WeGo · Add new application">
          <CallbackStateContextProvider
            callbackState={{
              page: authRedirectPage as PageRoute,
              state: {
                formData,
                updatedProfiles,
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
                  {formData.clusterAutomations.map(
                    (automation: ClusterAutomationFormData, index: number) => {
                      return (
                        <AppFields
                          key={index}
                          index={index}
                          formData={formData}
                          setFormData={setFormData}
                        />
                      );
                    },
                  )}
                </Grid>
                {automation.source?.kind === 'KindHelmRepository' ? (
                  <Profiles
                    // Temp fix to hide layers when using profiles in Add App until we update the BE
                    context="app"
                    isLoading={profilesIsLoading}
                    updatedProfiles={updatedProfiles}
                    setUpdatedProfiles={setUpdatedProfiles}
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
    profilesIsLoading,
    updatedProfiles,
    setUpdatedProfiles,
    showAuthDialog,
  ]);
};

export default () => (
  <ProfilesProvider>
    <AddApplication />
  </ProfilesProvider>
);
