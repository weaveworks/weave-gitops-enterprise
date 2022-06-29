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
import FormStepsNavigation from './Form/StepsNavigation';
import {
  Credential,
  ListProfileValuesResponse,
  UpdatedProfile,
} from '../../../types/custom';
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
import { TemplateObject } from '../../../cluster-services/cluster_services.pb';
import { useListConfig } from '../../../hooks/versions';

const large = weaveTheme.spacing.large;
const medium = weaveTheme.spacing.medium;
const base = weaveTheme.spacing.base;
const xxs = weaveTheme.spacing.xxs;
const xs = weaveTheme.spacing.xs;

const CredentialsWrapper = styled.div`
  display: flex;
  align-items: center;
  & .template-title {
    margin-right: ${medium};
  }
  & .credentials {
    display: flex;
    align-items: center;
    span {
      margin-right: ${xs};
    }
  }
  & .dropdown-toggle {
    border: 1px solid ${weaveTheme.colors.neutral10};
  }
  & .dropdown-popover {
    width: auto;
    flex-basis: content;
  }
  @media (max-width: 768px) {
    flex-direction: column;
    align-items: left;
    & .template-title {
      padding-bottom: ${base};
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
    PRPreview,
    setPRPreview,
    addCluster,
  } = useTemplates();
  const clustersCount = useClusters().count;
  const { data } = useListConfig();
  const repositoryURL = data?.repositoryURL || '';
  const { profiles, getProfileYaml } = useProfiles();
  const random = useMemo(() => Math.random().toString(36).substring(7), []);

  let initialFormData = {
    url: '',
    provider: '',
    branchName: `create-clusters-branch-${random}`,
    pullRequestTitle: 'Creates capi cluster',
    commitMessage: 'Creates capi cluster',
    pullRequestDescription: 'This PR creates a new cluster',
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
  const [steps, setSteps] = useState<string[]>([]);
  const [openPreview, setOpenPreview] = useState(false);
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const { templateName } = useParams<{ templateName: string }>();
  const history = useHistory();
  const [activeStep, setActiveStep] = useState<string | undefined>(undefined);
  const [clickedStep, setClickedStep] = useState<string>('');
  const isLargeScreen = useMediaQuery('(min-width:1632px)');
  const { setNotifications } = useNotifications();
  const authRedirectPage = `/clusters/templates/${activeTemplate?.name}/create`;

  const objectTitle = (object: TemplateObject, index: number) => {
    if (object.displayName && object.displayName !== '') {
      return `${index + 1}.${object.kind} (${object.displayName})`;
    }
    return `${index + 1}.${object.kind}`;
  };

  const handlePRPreview = useCallback(() => {
    setOpenPreview(true);
    const { url, provider, ...templateFields } = formData;
    renderTemplate({
      values: templateFields,
      credentials: infraCredential,
    });
  }, [formData, setOpenPreview, renderTemplate, infraCredential]);

  const getYaml = useCallback(
    (name: string, version: string) => {
      return getProfileYaml(name, version).then(
        (res: ListProfileValuesResponse) => res.message,
      );
    },
    [getProfileYaml],
  );

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
          profile.values.forEach(async value => {
            if (value.selected === true) {
              if (value.yaml === '') {
                value.yaml = await getYaml(profile.name, value.version);
              }
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
    [getYaml],
  );

  const handleAddCluster = useCallback(() => {
    const payload = {
      head_branch: formData.branchName,
      title: formData.pullRequestTitle,
      description: formData.pullRequestDescription,
      commit_message: formData.commitMessage,
      credentials: infraCredential,
      template_name: activeTemplate?.name,
      parameter_values: {
        ...formData,
      },
      values: encodedProfiles(selectedProfiles),
    };
    return addCluster(
      payload,
      getProviderToken(formData.provider as GitProvider),
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
      });
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
  ]);

  useEffect(() => {
    if (!activeTemplate) {
      clearCallbackState();
      setActiveTemplate(getTemplate(templateName));
    }

    const steps = activeTemplate?.objects?.map((object, index) =>
      objectTitle(object, index),
    );

    setSteps(steps as string[]);

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
              <Grid item xs={12} md={10}>
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
                    setActiveStep={setActiveStep}
                    clickedStep={clickedStep}
                    formData={formData}
                    setFormData={setFormData}
                    onFormDataUpdate={setFormData}
                    onPRPreview={handlePRPreview}
                  />
                ) : (
                  <Loader />
                )}
                {profiles.length > 0 && (
                  <Profiles
                    activeStep={activeStep}
                    setActiveStep={setActiveStep}
                    clickedStep={clickedStep}
                    selectedProfiles={selectedProfiles}
                    setSelectedProfiles={setSelectedProfiles}
                  />
                )}
                {openPreview && PRPreview ? (
                  <Preview
                    openPreview={openPreview}
                    setOpenPreview={setOpenPreview}
                    PRPreview={PRPreview}
                    activeStep={activeStep}
                    setActiveStep={setActiveStep}
                    clickedStep={clickedStep}
                  />
                ) : null}
                <GitOps
                  formData={formData}
                  setFormData={setFormData}
                  onSubmit={handleAddCluster}
                  activeStep={activeStep}
                  setActiveStep={setActiveStep}
                  clickedStep={clickedStep}
                  setClickedStep={setClickedStep}
                  showAuthDialog={showAuthDialog}
                  setShowAuthDialog={setShowAuthDialog}
                />
              </Grid>
              <Grid className={classes.steps} item md={2}>
                <FormStepsNavigation
                  steps={steps}
                  activeStep={activeStep}
                  setClickedStep={setClickedStep}
                  PRPreview={PRPreview}
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
    steps,
    activeStep,
    clickedStep,
    isLargeScreen,
    showAuthDialog,
    handlePRPreview,
    handleAddCluster,
    selectedProfiles,
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
