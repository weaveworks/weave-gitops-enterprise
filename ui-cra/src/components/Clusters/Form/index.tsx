import Divider from '@material-ui/core/Divider';
import Grid from '@material-ui/core/Grid';
import {
  createStyles,
  makeStyles,
  ThemeProvider,
} from '@material-ui/core/styles';
import useMediaQuery from '@material-ui/core/useMediaQuery';
import {
  Button,
  CallbackStateContextProvider,
  clearCallbackState,
  getProviderToken,
  LoadingPage,
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
  ClusterPRPreview,
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
const small = weaveTheme.spacing.small;

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
    previewCta: {
      display: 'flex',
      justifyContent: 'flex-end',
      padding: small,
      button: {
        width: '200px',
      },
    },
    previewLoading: {
      padding: base,
    },
  }),
);

function getInitialData(
  resource: any | undefined,
  callbackState: any,
  random: string,
) {
  const resourceData = resource && getCreateRequestAnnotation(resource);
  // update CLUSTER_NAME below
  const resourceName = resourceData?.parameter_values?.CLUSTER_NAME || '';
  const defaultFormData = {
    url: '',
    provider: '',
    branchName: resourceData
      ? `edit-resource-${resourceName}-branch-${random}`
      : `create-resource-branch-${random}`,
    pullRequestTitle: resourceData
      ? `Edits resource ${resourceName}`
      : 'Creates resource',
    commitMessage: resourceData
      ? `Edits resource ${resourceName}`
      : 'Creates resource',
    pullRequestDescription: resourceData
      ? 'This PR edits the resource'
      : 'This PR creates a new resource',
    parameterValues: resourceData?.parameter_values || {},
    clusterAutomations:
      resourceData?.kustomizations?.map((k: any) => ({
        name: k.metadata?.name,
        namespace: k.metadata?.namespace,
        path: k.spec?.path,
        target_namespace: k.spec?.target_namespace,
      })) || [],
  };

  const initialInfraCredentials = {
    ...resourceData?.infraCredential,
    ...callbackState?.state?.infraCredential,
  };

  const initialFormData = {
    ...defaultFormData,
    ...callbackState?.state?.formData,
  };

  return { initialFormData, initialInfraCredentials };
}

const getKustomizations = (formData: any) => {
  const { clusterAutomations } = formData;
  // filter out empty kustomization
  const filteredKustomizations = clusterAutomations.filter(
    (kustomization: any) => Object.values(kustomization).join('').trim() !== '',
  );
  return filteredKustomizations.map((kustomization: any): Kustomization => {
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
        createNamespace: kustomization.createNamespace,
      },
    };
  });
};

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
  const { parameterValues } = formData;
  return {
    headBranch: formData.branchName,
    title: formData.pullRequestTitle,
    description: formData.pullRequestDescription,
    commitMessage: formData.commitMessage,
    credentials: infraCredential,
    templateName,
    parameterValues,
    kustomizations: getKustomizations(formData),
    values: encodedProfiles(updatedProfiles),
  };
};

interface ResourceFormProps {
  resource?: any;
  template: TemplateEnriched;
}

const ResourceForm: FC<ResourceFormProps> = ({ template, resource }) => {
  // what type of template is it - templateType
  console.log(template);
  console.log(resource);
  const callbackState = useCallbackState();
  const classes = useStyles();
  const { renderTemplate, addCluster } = useTemplates();
  const { data } = useListConfig();
  const repositoryURL = data?.repositoryURL || '';
  const random = useMemo(() => Math.random().toString(36).substring(7), []);
  const { annotations, templateType } = template;

  const { initialFormData, initialInfraCredentials } = getInitialData(
    resource,
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
  const authRedirectPage = resource
    ? `/resources/${resource?.name}/edit`
    : `/templates/${template?.name}/create`;
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [PRPreview, setPRPreview] = useState<ClusterPRPreview | null>(null);
  const [loading, setLoading] = useState<boolean>(false);

  const handlePRPreview = useCallback(() => {
    const { url, provider, clusterAutomations, ...templateFields } = formData;
    setPreviewLoading(true);
    return renderTemplate(template.name, {
      values: templateFields.parameterValues,
      credentials: infraCredential,
      kustomizations: getKustomizations(formData),
    })
      .then(data => {
        setOpenPreview(true);
        setPRPreview(data);
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
    if (!resource) {
      setFormData((prevState: any) => ({
        ...prevState,
        pullRequestTitle: `Creates resource ${
          // update CLUSTER_NAME below
          formData.parameterValues.CLUSTER_NAME || ''
        }`,
      }));
    }
  }, [resource, formData.parameterValues, setFormData]);

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
              isCredentialEnabled={
                annotations?.['templates.weave.works/credentials-enabled']
              }
            />
          </CredentialsWrapper>
          <Divider
            className={!isLargeScreen ? classes.divider : classes.largeDivider}
          />
          <TemplateFields
            template={template}
            formData={formData}
            setFormData={setFormData}
          />
        </Grid>
        {/* Only show if resource kind is cluster? */}
        {templateType === 'cluster' && (
          <Profiles
            isLoading={profilesIsLoading}
            updatedProfiles={updatedProfiles}
            setUpdatedProfiles={setUpdatedProfiles}
            isProfilesEnabled={
              annotations?.['templates.weave.works/profiles-enabled']
            }
          />
        )}
        <Grid item xs={12} sm={10} md={10} lg={8}>
          {/* Only show if resource kind is cluster? */}
          <ApplicationsWrapper
            formData={formData}
            setFormData={setFormData}
            isKustomizationsEnabled={
              annotations?.['templates.weave.works/kustomizations-enabled']
            }
          />
          {previewLoading ? (
            <LoadingPage className={classes.previewLoading} />
          ) : (
            <div className={classes.previewCta}>
              <Button onClick={handlePRPreview}>PREVIEW PR</Button>
            </div>
          )}
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
    annotations,
  ]);
};

interface Props {
  template?: TemplateEnriched | null;
  resource?: any | null;
}

const ResourceFormWrapper: FC<Props> = ({ template, resource }) => {
  if (!template) {
    return (
      <Redirect
        to={{
          pathname: '/templates',
          state: {
            notification: [
              {
                message: {
                  text: 'No template information is available to create a resource.',
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
      {/* Only add Profiles Provider if resource kind is cluster? */}
      <ProfilesProvider cluster={resource || undefined} template={template}>
        <ResourceForm template={template} resource={resource || undefined} />
      </ProfilesProvider>
    </ThemeProvider>
  );
};

export default ResourceFormWrapper;
