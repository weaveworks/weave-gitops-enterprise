import React, { FC, useCallback, useEffect, useMemo, useState } from 'react';
import useTemplates from '../../../contexts/Templates';
import useClusters from '../../../contexts/Clusters';
import useNotifications from '../../../contexts/Notifications';
import useVersions from '../../../contexts/Versions';
import useProfiles from '../../../contexts/Profiles';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';
import { useParams } from 'react-router-dom';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import Grid from '@material-ui/core/Grid';
import Divider from '@material-ui/core/Divider';
import { useHistory } from 'react-router-dom';
import FormStepsNavigation from './Form/StepsNavigation';
import {
  Credential,
  TemplateObject,
  UpdatedProfile,
} from '../../../types/custom';
import styled from 'styled-components';
import useMediaQuery from '@material-ui/core/useMediaQuery';
import CredentialsProvider from '../../../contexts/Credentials/Provider';
import { Loader } from '../../Loader';
import {
  CallbackStateContextProvider,
  clearCallbackState,
  getCallbackState,
  getProviderToken,
} from '@weaveworks/weave-gitops';
import { isUnauthenticated } from '../../../utils/request';
import Compose from '../../ProvidersCompose';
import TemplateFields from './Form/Partials/TemplateFields';
import Credentials from './Form/Partials/Credentials';
import GitOps from './Form/Partials/GitOps';
import Preview from './Form/Partials/Preview';
import ProfilesProvider from '../../../contexts/Profiles/Provider';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import Profiles from './Form/Partials/Profiles';

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
    border: 1px solid #e5e5e5;
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
    creatingPR,
    addCluster,
    setError,
  } = useTemplates();
  const clustersCount = useClusters().count;
  const { repositoryURL } = useVersions();
  const { updatedProfiles } = useProfiles();
  const random = useMemo(() => Math.random().toString(36).substring(7), []);

  let initialFormData = {
    url: '',
    provider: '',
    branchName: `create-clusters-branch`,
    pullRequestTitle: 'Creates capi cluster',
    commitMessage: 'Creates capi cluster',
    pullRequestDescription: 'This PR creates a new cluster',
  };

  let initialProfiles = updatedProfiles;

  let initialInfraCredential = {} as Credential;

  const callbackState = getCallbackState();

  if (callbackState) {
    initialFormData = {
      ...initialFormData,
      ...callbackState.state.formData,
    };
    initialProfiles = [...initialProfiles, ...callbackState.state.profiles];
    initialInfraCredential = {
      ...initialInfraCredential,
      ...callbackState.state.infraCredential,
    };
  }

  const [formData, setFormData] = useState<any>(initialFormData);
  const [profiles, setProfiles] = useState<UpdatedProfile[]>(initialProfiles);
  const [infraCredential, setInfraCredential] = useState<Credential>(
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

  const encodedProfiles = useCallback(
    (profiles: UpdatedProfile[]) =>
      profiles.reduce(
        (
          accumulator: {
            name: string;
            version: string;
            values: string;
          }[],
          profile,
        ) => {
          profile.values.forEach(value => {
            if (value.selected === true)
              accumulator.push({
                name: profile.name,
                version: value.version,
                values: btoa(value.yaml),
              });
          });
          return accumulator;
        },
        [],
      ),
    [],
  );

  const handleAddCluster = useCallback(
    () =>
      addCluster(
        {
          head_branch: formData.branchName,
          title: formData.pullRequestTitle,
          description: formData.pullRequestDescription,
          commit_message: formData.commitMessage,
          credentials: infraCredential,
          template_name: activeTemplate?.name,
          parameter_values: {
            ...formData,
          },
          values: encodedProfiles(profiles),
        },
        getProviderToken(formData.provider as GitProvider),
      )
        .then(() => {
          setPRPreview(null);
          history.push('/clusters');
        })
        .catch(error => {
          if (isUnauthenticated(error.code)) {
            setShowAuthDialog(true);
          } else {
            setNotifications([{ message: error.message, variant: 'danger' }]);
          }
        }),
    [
      profiles,
      addCluster,
      formData,
      activeTemplate?.name,
      infraCredential,
      history,
      setNotifications,
      encodedProfiles,
      setPRPreview,
    ],
  );

  useEffect(() => {
    if (!activeTemplate) {
      setActiveTemplate(getTemplate(templateName));
    }

    const steps = activeTemplate?.objects?.map((object, index) =>
      objectTitle(object, index),
    );

    setSteps(steps as string[]);

    return history.listen(() => {
      setActiveTemplate(null);
      setPRPreview(null);
      setError(null);
      clearCallbackState();
    });
  }, [
    activeTemplate,
    getTemplate,
    setActiveTemplate,
    templateName,
    history,
    setError,
    setPRPreview,
  ]);

  useEffect(() => {
    if (!callbackState) {
      setFormData((prevState: any) => ({
        ...prevState,
        url: repositoryURL,
        branchName: `create-clusters-branch-${random}`,
      }));
    }
  }, [callbackState, infraCredential, random, repositoryURL]);

  return useMemo(() => {
    return (
      <PageTemplate documentTitle="WeGo · Create new cluster">
        <CallbackStateContextProvider
          callbackState={{
            page: authRedirectPage as PageRoute,
            state: { infraCredential, formData, profiles },
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
            <Grid container spacing={2}>
              <Grid item xs={12} md={9}>
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
                <TemplateFields
                  activeTemplate={activeTemplate}
                  setActiveStep={setActiveStep}
                  clickedStep={clickedStep}
                  formData={formData}
                  setFormData={setFormData}
                  onFormDataUpdate={setFormData}
                  onPRPreview={handlePRPreview}
                />
                <Profiles
                  activeStep={activeStep}
                  setActiveStep={setActiveStep}
                  clickedStep={clickedStep}
                  profiles={profiles}
                  setProfiles={setProfiles}
                />
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
                {creatingPR && <Loader />}
              </Grid>
              <Grid className={classes.steps} item md={3}>
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
    profiles,
    infraCredential,
    activeTemplate,
    clustersCount,
    classes,
    openPreview,
    PRPreview,
    creatingPR,
    steps,
    activeStep,
    clickedStep,
    isLargeScreen,
    showAuthDialog,
    handlePRPreview,
    handleAddCluster,
  ]);
};

const AddClusterWithCredentials = () => (
  <Compose components={[ProfilesProvider, CredentialsProvider]}>
    <AddCluster />
  </Compose>
);

export default AddClusterWithCredentials;
