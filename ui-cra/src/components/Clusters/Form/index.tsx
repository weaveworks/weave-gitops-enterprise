import Divider from '@material-ui/core/Divider';
import Grid from '@material-ui/core/Grid';
import {
  createStyles,
  makeStyles,
  ThemeProvider,
} from '@material-ui/core/styles';
import useMediaQuery from '@material-ui/core/useMediaQuery';
import {
  CallbackStateContextProvider,
  clearCallbackState,
  getProviderToken,
  theme as weaveTheme,
} from '@weaveworks/weave-gitops';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import _ from 'lodash';
import React, { FC, useCallback, useEffect, useMemo, useState } from 'react';
import { useHistory, Redirect } from 'react-router-dom';
import styled from 'styled-components';
import {
  CreatePullRequestRequest,
  Kustomization,
  ProfileValues,
} from '../../../cluster-services/cluster_services.pb';
import useNotifications from '../../../contexts/Notifications';
import useProfiles from '../../../contexts/Profiles';
import ProfilesProvider from '../../../contexts/Profiles/Provider';
import useTemplates from '../../../hooks/templates';
import { useListConfig } from '../../../hooks/versions';
import { localEEMuiTheme } from '../../../muiTheme';
import {
  Credential,
  GitopsClusterEnriched,
  ProfilesIndex,
  TemplateEnriched,
} from '../../../types/custom';
import { utf8_to_b64 } from '../../../utils/base64';
import { useCallbackState } from '../../../utils/callback-state';
import {
  FLUX_BOOSTRAP_KUSTOMIZATION_NAME,
  FLUX_BOOSTRAP_KUSTOMIZATION_NAMESPACE,
} from '../../../utils/config';
import { isUnauthenticated, removeToken } from '../../../utils/request';
import { ApplicationsWrapper } from './Partials/ApplicationsWrapper';
import Credentials from './Partials/Credentials';
import GitOps from './Partials/GitOps';
import Preview from './Partials/Preview';
import Profiles from './Partials/Profiles';
import TemplateFields from './Partials/TemplateFields';
import { getCreateRequestAnnotation } from './utils';

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

function getInitialData(
  cluster: GitopsClusterEnriched | undefined,
  callbackState: any,
  random: string,
) {
  const clusterData = cluster && getCreateRequestAnnotation(cluster);
  const clusterName = clusterData?.parameter_values?.CLUSTER_NAME || '';
  const defaultFormData = {
    url: '',
    provider: '',
    branchName: clusterData
      ? `edit-cluster-${clusterName}-branch-${random}`
      : `create-clusters-branch-${random}`,
    pullRequestTitle: clusterData
      ? `Edits cluster ${clusterName}`
      : 'Creates cluster',
    commitMessage: clusterData
      ? `Edits capi cluster ${clusterName}`
      : 'Creates capi cluster',
    pullRequestDescription: clusterData
      ? 'This PR edits the cluster'
      : 'This PR creates a new cluster',
    parameterValues: clusterData?.parameter_values || {},
    clusterAutomations:
      clusterData?.kustomizations?.map((k: any) => ({
        name: k.metadata?.name,
        namespace: k.metadata?.namespace,
        path: k.spec?.path,
        target_namespace: k.spec?.target_namespace,
      })) || [],
  };

  const initialInfraCredentials = {
    ...clusterData?.infraCredential,
    ...callbackState?.state?.infraCredential,
  };

  const initialFormData = {
    ...defaultFormData,
    ...callbackState?.state?.formData,
  };

  return { initialFormData, initialInfraCredentials };
}

const encodedProfiles = (profiles: ProfilesIndex): ProfileValues[] =>
  _.sortBy(Object.values(profiles), 'name')
    .filter(p => p.selected)
    .map(p => {
      // FIXME: handle this somehow..
      const v = p.values.find(v => v.selected)!;
      return {
        name: p.name,
        version: v.version,
        values: utf8_to_b64(v.yaml),
        layer: p.layer,
        namespace: p.namespace,
      };
    });

const toPayload = (
  formData: any,
  infraCredential: any,
  templateName: string,
  updatedProfiles: ProfilesIndex,
): CreatePullRequestRequest => {
  const { clusterAutomations, parameterValues } = formData;
  // filter out empty kustomization
  const filteredKustomizations = clusterAutomations.filter(
    (kustomization: any) => Object.values(kustomization).join('').trim() !== '',
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
            name: FLUX_BOOSTRAP_KUSTOMIZATION_NAME,
            namespace: FLUX_BOOSTRAP_KUSTOMIZATION_NAMESPACE,
          },
          targetNamespace: kustomization.target_namespace,
        },
      };
    },
  );
  return {
    headBranch: formData.branchName,
    title: formData.pullRequestTitle,
    description: formData.pullRequestDescription,
    commitMessage: formData.commitMessage,
    credentials: infraCredential,
    templateName,
    parameterValues,
    kustomizations,
    values: encodedProfiles(updatedProfiles),
  };
};

interface ClusterFormProps {
  cluster?: GitopsClusterEnriched;
  template: TemplateEnriched;
}

const ClusterForm: FC<ClusterFormProps> = ({ template, cluster }) => {
  const callbackState = useCallbackState();
  const classes = useStyles();
  const { renderTemplate, addCluster } = useTemplates();
  const { data } = useListConfig();
  const repositoryURL = data?.repositoryURL || '';
  const random = useMemo(() => Math.random().toString(36).substring(7), []);

  const { initialFormData, initialInfraCredentials } = getInitialData(
    cluster,
    callbackState,
    random,
  );
  const [formData, setFormData] = useState<any>(initialFormData);
  const [infraCredential, setInfraCredential] = useState<Credential | null>(
    initialInfraCredentials,
  );
  const { profiles, isLoading: profilesIsLoading } = useProfiles();
  const [updatedProfiles, setUpdatedProfiles] = useState<ProfilesIndex>({});

  useEffect(() => {
    clearCallbackState();
  }, []);

  useEffect(() => {
    setUpdatedProfiles({
      ..._.keyBy(profiles, 'name'),
      ...callbackState?.state?.updatedProfiles,
    });
  }, [callbackState?.state?.updatedProfiles, profiles]);

  const [openPreview, setOpenPreview] = useState(false);
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const history = useHistory();
  const isLargeScreen = useMediaQuery('(min-width:1632px)');
  const { setNotifications } = useNotifications();
  const authRedirectPage = cluster
    ? `/clusters/${cluster?.name}/edit`
    : `/templates/${template?.name}/create`;
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [PRPreview, setPRPreview] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(false);

  const handlePRPreview = useCallback(() => {
    const { url, provider, clusterAutomations, ...templateFields } = formData;
    setPreviewLoading(true);
    return renderTemplate(template.name, {
      values: templateFields.parameterValues,
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
    template.name,
  ]);

  const handleAddCluster = useCallback(() => {
    const payload = toPayload(
      formData,
      infraCredential,
      template.name,
      updatedProfiles,
    );
    setLoading(true);
    return addCluster(
      payload,
      getProviderToken(formData.provider as GitProvider),
      template.templateKind,
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
    updatedProfiles,
    addCluster,
    formData,
    infraCredential,
    history,
    setNotifications,
    setPRPreview,
    template.name,
    template.templateKind,
  ]);

  useEffect(() => {
    setFormData((prevState: any) => ({
      ...prevState,
      url: repositoryURL,
    }));
  }, [repositoryURL]);

  useEffect(() => {
    if (!cluster) {
      setFormData((prevState: any) => ({
        ...prevState,
        pullRequestTitle: `Creates cluster ${
          formData.parameterValues.CLUSTER_NAME || ''
        }`,
      }));
    }
  }, [cluster, formData.parameterValues, setFormData]);

  return useMemo(() => {
    return (
      <CallbackStateContextProvider
        callbackState={{
          page: authRedirectPage as PageRoute,
          state: {
            infraCredential,
            formData,
            updatedProfiles,
          },
        }}
      >
        <Grid item xs={12} sm={10} md={10} lg={8}>
          <CredentialsWrapper>
            <div className="template-title">
              Template: <span>{template.name}</span>
            </div>
            <Credentials
              infraCredential={infraCredential}
              setInfraCredential={setInfraCredential}
            />
          </CredentialsWrapper>

          <Divider
            className={!isLargeScreen ? classes.divider : classes.largeDivider}
          />
          <TemplateFields
            template={template}
            formData={formData}
            setFormData={setFormData}
            onPRPreview={handlePRPreview}
            previewLoading={previewLoading}
          />
        </Grid>
        <Profiles
          isLoading={profilesIsLoading}
          updatedProfiles={updatedProfiles}
          setUpdatedProfiles={setUpdatedProfiles}
        />
        <Grid item xs={12} sm={10} md={10} lg={8}>
          <ApplicationsWrapper formData={formData} setFormData={setFormData} />
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
      </CallbackStateContextProvider>
    );
  }, [
    authRedirectPage,
    template,
    formData,
    infraCredential,
    classes,
    openPreview,
    PRPreview,
    profilesIsLoading,
    isLargeScreen,
    showAuthDialog,
    setUpdatedProfiles,
    handlePRPreview,
    handleAddCluster,
    updatedProfiles,
    previewLoading,
    loading,
  ]);
};

interface Props {
  template?: TemplateEnriched | null;
  cluster?: GitopsClusterEnriched | null;
}

const ClusterFormWrapper: FC<Props> = ({ template, cluster }) => {
  if (!template) {
    return (
      <Redirect
        to={{
          pathname: '/templates',
          state: {
            notification: [
              {
                message: {
                  text: 'No template information is available to create a cluster.',
                },
                variant: 'danger',
              },
            ],
          },
        }}
      />
    );
  }

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <ProfilesProvider cluster={cluster || undefined} template={template}>
        <ClusterForm template={template} cluster={cluster || undefined} />
      </ProfilesProvider>
    </ThemeProvider>
  );
};

export default ClusterFormWrapper;
