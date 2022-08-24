import React, { FC, useCallback, useEffect, useMemo, useState } from 'react';
import { ThemeProvider } from '@material-ui/core/styles';
import useTemplates from '../../../contexts/Templates';
import useClusters from '../../../contexts/Clusters';
import useNotifications from '../../../contexts/Notifications';
import useProfiles from '../../../contexts/Profiles';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';
import { useParams } from 'react-router-dom';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import Grid from '@material-ui/core/Grid';
import Divider from '@material-ui/core/Divider';
import { useHistory } from 'react-router-dom';
import { Credential, UpdatedProfile } from '../../../types/custom';
import styled from 'styled-components';
import useMediaQuery from '@material-ui/core/useMediaQuery';
import { Loader } from '../../Loader';
import {
  CallbackStateContextProvider,
  clearCallbackState,
  getCallbackState,
  getProviderToken,
} from '@weaveworks/weave-gitops';
import { isUnauthenticated, removeToken } from '../../../utils/request';
import TemplateFields from './Form/Partials/TemplateFields';
import Credentials from './Form/Partials/Credentials';
import GitOps from './Form/Partials/GitOps';
import Preview from './Form/Partials/Preview';
import ProfilesProvider from '../../../contexts/Profiles/Provider';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import Profiles from './Form/Partials/Profiles';
import { localEEMuiTheme } from '../../../muiTheme';
import { useListConfig } from '../../../hooks/versions';
import { ApplicationsWrapper } from './Form/Partials/ApplicationsWrapper';
import { Kustomization } from '../../../cluster-services/cluster_services.pb';

const large = weaveTheme.spacing.large;
const medium = weaveTheme.spacing.medium;
const base = weaveTheme.spacing.base;
const xxs = weaveTheme.spacing.xxs;

const CredentialsWrapper = styled.div`
  display: flex;
  align-items: center;
  & .template-title {
    margin-right: ${({ theme }) => theme.spacing.medium};
  }
  & .credentials {
    display: flex;
    align-items: center;
    span {
      margin-right: ${({ theme }) => theme.spacing.xs};
    }
  }
  & .dropdown-toggle {
    border: 1px solid ${({ theme }) => theme.colors.neutral10};
  }
  & .dropdown-popover {
    width: auto;
    flex-basis: content;
  }
  @media (max-width: 768px) {
    flex-direction: column;
    align-items: left;
    & .template-title {
      padding-bottom: ${({ theme }) => theme.spacing.base};
    }
  }
`;

const useStyles = makeStyles(theme =>
  createStyles({
    divider: {
      marginTop: medium,
      marginBottom: base,
    },
    largeDivider: {
      margin: `${large} 0`,
    },
    steps: {
      display: 'flex',
      flexDirection: 'column',
      [theme.breakpoints.down('md')]: {
        visibility: 'hidden',
        height: 0,
      },
      paddingRight: xxs,
    },
  }),
);

const AddCluster: FC = () => {
  const classes = useStyles();
  const {
    getTemplate,
    activeTemplate,
    setActiveTemplate,
    renderTemplate,
    addCluster,
  } = useTemplates();
  const clustersCount = useClusters().count;
  const { data } = useListConfig();
  const repositoryURL = data?.repositoryURL || '';
  const { profiles } = useProfiles();
  const random = useMemo(() => Math.random().toString(36).substring(7), []);

  let initialFormData = {
    url: '',
    provider: '',
    branchName: `create-clusters-branch-${random}`,
    pullRequestTitle: 'Creates cluster',
    commitMessage: 'Creates capi cluster',
    pullRequestDescription: 'This PR creates a new cluster',
    clusterAutomations: [],
  };

  let initialProfiles = [] as UpdatedProfile[];

  let initialInfraCredential = {} as Credential;

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
    initialInfraCredential = {
      ...initialInfraCredential,
      ...callbackState.state.infraCredential,
    };
  }

  const [formData, setFormData] = useState<any>(initialFormData);
  const [selectedProfiles, setSelectedProfiles] =
    useState<UpdatedProfile[]>(initialProfiles);
  const [infraCredential, setInfraCredential] = useState<Credential | null>(
    initialInfraCredential,
  );
  const [openPreview, setOpenPreview] = useState(false);
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const { templateName } = useParams<{ templateName: string }>();
  const history = useHistory();
  const isLargeScreen = useMediaQuery('(min-width:1632px)');
  const { setNotifications } = useNotifications();
  const authRedirectPage = `/clusters/templates/${activeTemplate?.name}/create`;
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [PRPreview, setPRPreview] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(false);

  const handlePRPreview = useCallback(() => {
    const { url, provider, clusterAutomations, ...templateFields } = formData;
    setPreviewLoading(true);
    return renderTemplate({
      values: templateFields,
      credentials: infraCredential,
    })
      .then(data => {
        setOpenPreview(true);
        setPRPreview(data.renderedTemplate);
      })
      .catch(err =>
        setNotifications([
          { message: { text: err.message }, variant: 'danger' },
        ]),
      )
      .finally(() => setPreviewLoading(false));
  }, [
    formData,
    setOpenPreview,
    renderTemplate,
    infraCredential,
    setNotifications,
  ]);

  const encodedProfiles = useCallback(
    (profiles: UpdatedProfile[]) =>
      profiles.reduce(
        (
          accumulator: {
            name: string;
            version: string;
            values: string;
            layer?: string;
            namespace?: string;
          }[],
          profile,
        ) => {
          profile.values.forEach(value => {
            if (value.selected === true) {
              accumulator.push({
                name: profile.name,
                version: value.version,
                values: btoa(value.yaml),
                layer: profile.layer,
                namespace: profile.namespace,
              });
            }
          });
          return accumulator;
        },
        [],
      ),
    [],
  );

  const handleAddCluster = useCallback(() => {
    const { clusterAutomations, ...rest } = formData;
    // filter out empty kustomization
    const filteredKustomizations = clusterAutomations.filter(
      (kustomization: any) =>
        Object.values(kustomization).join('').trim() !== '',
    );
    const kustomizations = filteredKustomizations.map(
      (kustomization: any): Kustomization => {
        return {
          metadata: {
            name: kustomization.name,
            namespace: kustomization.namespace,
          },
          spec: {
            path: kustomization.path,
            sourceRef: {
              name: 'flux-system',
              namespace: 'flux-system',
            },
          },
        };
      },
    );
    const payload = {
      head_branch: formData.branchName,
      title: formData.pullRequestTitle,
      description: formData.pullRequestDescription,
      commit_message: formData.commitMessage,
      credentials: infraCredential,
      template_name: activeTemplate?.name,
      parameter_values: {
        ...rest,
      },
      kustomizations,
      values: encodedProfiles(selectedProfiles),
    };
    setLoading(true);
    return addCluster(
      payload,
      getProviderToken(formData.provider as GitProvider),
      activeTemplate?.templateKind || '',
    )
      .then(response => {
        setPRPreview(null);
        history.push('/clusters');
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
  }, [
    selectedProfiles,
    addCluster,
    formData,
    activeTemplate?.name,
    infraCredential,
    history,
    setNotifications,
    encodedProfiles,
    setPRPreview,
    activeTemplate?.templateKind,
  ]);

  useEffect(() => {
    if (!activeTemplate) {
      clearCallbackState();
      setActiveTemplate(getTemplate(templateName));
    }

    return history.listen(() => {
      setActiveTemplate(null);
      setPRPreview(null);
    });
  }, [
    activeTemplate,
    getTemplate,
    setActiveTemplate,
    templateName,
    history,
    setPRPreview,
  ]);

  useEffect(() => {
    if (!callbackState) {
      setFormData((prevState: any) => ({
        ...prevState,
        url: repositoryURL,
      }));
    }
  }, [callbackState, infraCredential, repositoryURL, profiles]);

  useEffect(() => {
    setFormData((prevState: any) => ({
      ...prevState,
      pullRequestTitle: `Creates cluster ${formData.CLUSTER_NAME || ''}`,
    }));
  }, [formData.CLUSTER_NAME, setFormData]);

  return useMemo(() => {
    return (
      <PageTemplate documentTitle="WeGo · Create new cluster">
        <CallbackStateContextProvider
          callbackState={{
            page: authRedirectPage as PageRoute,
            state: {
              infraCredential,
              formData,
              selectedProfiles,
            },
          }}
        >
          <SectionHeader
            className="count-header"
            path={[
              { label: 'Clusters', url: '/', count: clustersCount },
              { label: 'Create new cluster' },
            ]}
          />
          <ContentWrapper>
            <Grid container>
              <Grid item xs={12} sm={10} md={10} lg={8}>
                <Title>Create new cluster with template</Title>
                <CredentialsWrapper>
                  <div className="template-title">
                    Template: <span>{activeTemplate?.name}</span>
                  </div>
                  <Credentials
                    infraCredential={infraCredential}
                    setInfraCredential={setInfraCredential}
                  />
                </CredentialsWrapper>

                <Divider
                  className={
                    !isLargeScreen ? classes.divider : classes.largeDivider
                  }
                />
                {activeTemplate ? (
                  <TemplateFields
                    activeTemplate={activeTemplate}
                    formData={formData}
                    setFormData={setFormData}
                    onFormDataUpdate={setFormData}
                    onPRPreview={handlePRPreview}
                    previewLoading={previewLoading}
                  />
                ) : (
                  <Loader />
                )}
              </Grid>
              {profiles.length > 0 && (
                <Profiles
                  context="app"
                  selectedProfiles={selectedProfiles}
                  setSelectedProfiles={setSelectedProfiles}
                />
              )}
              <Grid item xs={12} sm={10} md={10} lg={8}>
                {
                  <ApplicationsWrapper
                    formData={formData}
                    setFormData={setFormData}
                  ></ApplicationsWrapper>
                }
              </Grid>

              {openPreview && PRPreview ? (
                <Preview
                  openPreview={openPreview}
                  setOpenPreview={setOpenPreview}
                  PRPreview={PRPreview}
                />
              ) : null}
              <Grid item xs={12} sm={10} md={10} lg={8}>
                <GitOps
                  loading={loading}
                  formData={formData}
                  setFormData={setFormData}
                  onSubmit={handleAddCluster}
                  showAuthDialog={showAuthDialog}
                  setShowAuthDialog={setShowAuthDialog}
                />
              </Grid>
            </Grid>
          </ContentWrapper>
        </CallbackStateContextProvider>
      </PageTemplate>
    );
  }, [
    authRedirectPage,
    formData,
    profiles.length,
    infraCredential,
    activeTemplate,
    clustersCount,
    classes,
    openPreview,
    PRPreview,
    isLargeScreen,
    showAuthDialog,
    handlePRPreview,
    handleAddCluster,
    selectedProfiles,
    previewLoading,
    loading,
  ]);
};

const AddClusterWithCredentials = () => (
  <ThemeProvider theme={localEEMuiTheme}>
    <ProfilesProvider>
      <AddCluster />
    </ProfilesProvider>
  </ThemeProvider>
);

export default AddClusterWithCredentials;
