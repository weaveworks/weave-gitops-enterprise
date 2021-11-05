import React, { FC, useCallback, useEffect, useMemo, useState } from 'react';
import useTemplates from '../../../contexts/Templates';
import useClusters from '../../../contexts/Clusters';
import useNotifications from '../../../contexts/Notifications';
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
  getProviderToken,
  GithubDeviceAuthModal,
} from '@weaveworks/weave-gitops';
import { isUnauthenticated } from '../../../utils/request';
import Compose from '../../ProvidersCompose';
import TemplateFields from './Form/Partials/TemplateFields';
import ProfilesProvider from '../../../contexts/Profiles/Provider';
import Credentials from './Form/Partials/Credentials';
import GitOps from './Form/Partials/GitOps';
import Preview from './Form/Partials/Preview';

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
  const [formData, setFormData] = useState({});
  const [updatedProfiles, setUpdatedProfiles] = useState<UpdatedProfile[]>([]);
  const [steps, setSteps] = useState<string[]>([]);
  const [openPreview, setOpenPreview] = useState(false);
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const { templateName } = useParams<{ templateName: string }>();
  const history = useHistory();
  const [activeStep, setActiveStep] = useState<string | undefined>(undefined);
  const [clickedStep, setClickedStep] = useState<string>('');
  const [infraCredential, setInfraCredential] =
    useState<Credential | null>(null);
  const isLargeScreen = useMediaQuery('(min-width:1632px)');
  const { setNotifications } = useNotifications();

  const objectTitle = (object: TemplateObject, index: number) => {
    if (object.displayName && object.displayName !== '') {
      return `${index + 1}.${object.kind} (${object.displayName})`;
    }
    return `${index + 1}.${object.kind}`;
  };

  const handlePRPreview = useCallback(() => {
    setOpenPreview(true);
    setClickedStep('Preview');
    renderTemplate({ values: formData, credentials: infraCredential });
  }, [formData, setOpenPreview, renderTemplate, infraCredential]);

  const encodedProfiles = useCallback(
    (profiles: UpdatedProfile[]) =>
      profiles?.map(profile => {
        return {
          name: profile.name,
          version: profile.version,
          values: btoa(profile.values),
        };
      }),
    [],
  );

  const handleAddCluster = useCallback(
    (gitOps: {
      head_branch: string;
      title: string;
      description: string;
      commit_message: string;
    }) =>
      addCluster(
        {
          ...gitOps,
          credentials: infraCredential,
          template_name: activeTemplate?.name,
          parameter_values: {
            ...formData,
          },
          values: encodedProfiles(updatedProfiles),
        },
        getProviderToken('github'),
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
      addCluster,
      formData,
      activeTemplate?.name,
      infraCredential,
      history,
      setNotifications,
      encodedProfiles,
      setPRPreview,
      updatedProfiles,
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

  return useMemo(() => {
    return (
      <PageTemplate documentTitle="WeGo · Create new cluster">
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
                <Credentials onSelect={setInfraCredential} />
              </CredentialsWrapper>
              <Divider
                className={
                  !isLargeScreen ? classes.divider : classes.largeDivider
                }
              />
              <TemplateFields
                activeTemplate={activeTemplate}
                activeStep={activeStep}
                setActiveStep={setActiveStep}
                clickedStep={clickedStep}
                formData={formData}
                onFormDataUpdate={setFormData}
                onProfilesUpdate={setUpdatedProfiles}
                onPRPreview={handlePRPreview}
              />
              {openPreview ? (
                <>
                  {PRPreview ? (
                    <>
                      <Preview
                        PRPreview={PRPreview}
                        activeStep={activeStep}
                        setActiveStep={setActiveStep}
                        clickedStep={clickedStep}
                      />
                      <GitOps
                        onSubmit={handleAddCluster}
                        activeStep={activeStep}
                        setActiveStep={setActiveStep}
                        clickedStep={clickedStep}
                        setClickedStep={setClickedStep}
                      />
                      {creatingPR && <Loader />}
                    </>
                  ) : (
                    <Loader />
                  )}
                </>
              ) : null}
              <GithubDeviceAuthModal
                onClose={() => setShowAuthDialog(false)}
                onSuccess={() => {
                  setShowAuthDialog(false);
                  setNotifications([
                    {
                      message:
                        'Authentication completed successfully. Please proceed with creating the PR.',
                      variant: 'success',
                    },
                  ]);
                }}
                open={showAuthDialog}
                repoName="config"
              />
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
      </PageTemplate>
    );
  }, [
    formData,
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
    setNotifications,
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
