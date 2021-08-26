import React, {
  ChangeEvent,
  FC,
  FormEvent,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react';
import useTemplates from '../../../contexts/Templates';
import useClusters from '../../../contexts/Clusters';
import useCredentials from '../../../contexts/Credentials';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { useParams } from 'react-router-dom';
import Form from '@rjsf/material-ui';
import { JSONSchema7 } from 'json-schema';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { Button, Dropdown, DropdownItem } from 'weaveworks-ui-components';
import { ISubmitEvent, ObjectFieldTemplateProps } from '@rjsf/core';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import Grid from '@material-ui/core/Grid';
import { Input } from '../../../utils/form';
import { Loader } from '../../Loader';
import Divider from '@material-ui/core/Divider';
import { useHistory } from 'react-router-dom';
import * as Grouped from './Form/GroupedSchema';
import * as UiTemplate from './Form/UITemplate';
import FormSteps, { FormStep } from './Form/Steps';
import FormStepsNavigation from './Form/StepsNavigation';
import { Credential } from '../../../types/custom';
import styled from 'styled-components';
import useMediaQuery from '@material-ui/core/useMediaQuery';
import CredentialsProvider from '../../../contexts/Credentials/Provider';

const large = weaveTheme.spacing.large;
const medium = weaveTheme.spacing.medium;
const base = weaveTheme.spacing.base;
const xxs = weaveTheme.spacing.xxs;
const xs = weaveTheme.spacing.xs;
const small = weaveTheme.spacing.small;

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
    form: {
      paddingTop: base,
    },
    create: {
      paddingTop: small,
    },
    title: {
      fontSize: weaveTheme.fontSizes.large,
      fontWeight: 600,
      paddingBottom: weaveTheme.spacing.medium,
      color: weaveTheme.colors.gray600,
    },
    divider: {
      marginTop: medium,
      marginBottom: base,
    },
    largeDivider: {
      margin: `${large} 0`,
    },
    textarea: {
      width: '100%',
      padding: xs,
      border: '1px solid #E5E5E5',
    },
    previewCTA: {
      display: 'flex',
      justifyContent: 'flex-end',
      paddingTop: small,
      paddingBottom: base,
    },
    createCTA: {
      display: 'flex',
      justifyContent: 'center',
      paddingTop: base,
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
    errorMessage: {
      margin: `${weaveTheme.spacing.medium} 0`,
      padding: weaveTheme.spacing.small,
      border: '1px solid #E6E6E6',
      borderRadius: weaveTheme.borderRadius.soft,
      color: weaveTheme.colors.orange600,
    },
  }),
);

const AddCluster: FC = () => {
  const classes = useStyles();
  const { credentials, loading, getCredential } = useCredentials();
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
  const random = Math.random().toString(36).substring(7);
  const clustersCount = useClusters().count;
  const [formData, setFormData] = useState({});
  const [steps, setSteps] = useState<string[]>([]);
  const [openPreview, setOpenPreview] = useState(false);
  const [branchName, setBranchName] = useState<string>(
    `create-clusters-branch-${random}`,
  );
  const [pullRequestTitle, setPullRequestTitle] = useState<string>(
    'Creates capi cluster',
  );
  const [commitMessage, setCommitMessage] = useState<string>(
    'Creates capi cluster',
  );
  const [pullRequestDescription, setPullRequestDescription] = useState<string>(
    'This PR creates a new cluster',
  );

  const rows = (PRPreview?.split('\n').length || 0) - 1;
  const { templateName } = useParams<{ templateName: string }>();
  const history = useHistory();
  const [activeStep, setActiveStep] = useState<string | undefined>(undefined);
  const [clickedStep, setClickedStep] = useState<string>('');
  const [infraCredential, setInfraCredential] =
    useState<Credential | null>(null);
  const isLargeScreen = useMediaQuery('(min-width:1632px)');

  const credentialsItems: DropdownItem[] = useMemo(
    () => [
      ...credentials.map((credential: Credential) => {
        const { kind, namespace, name } = credential;
        return {
          label: `${kind}/${namespace || 'default'}/${name}`,
          value: name || '',
        };
      }),
      { label: 'None', value: '' },
    ],
    [credentials],
  );

  const handleSelectCredentials = useCallback(
    (event: FormEvent<HTMLInputElement>, value: string) => {
      const credential = getCredential(value);
      setInfraCredential(credential);
    },
    [getCredential],
  );

  const handlePreview = useCallback(
    (event: ISubmitEvent<any>) => {
      setFormData(event.formData);
      setOpenPreview(true);
      setClickedStep('Preview');
      renderTemplate({ values: event.formData, credentials: infraCredential });
    },
    [setOpenPreview, setFormData, renderTemplate, infraCredential],
  );

  const handleChangeBranchName = useCallback(
    (event: ChangeEvent<HTMLInputElement>) => setBranchName(event.target.value),
    [],
  );

  const handleChangePullRequestTitle = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setPullRequestTitle(event.target.value),
    [],
  );

  const handleChangeCommitMessage = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setCommitMessage(event.target.value),
    [],
  );

  const handleChangePRDescription = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setPullRequestDescription(event.target.value),
    [],
  );

  const handleAddCluster = useCallback(() => {
    addCluster({
      credentials: infraCredential,
      head_branch: branchName,
      title: pullRequestTitle,
      description: pullRequestDescription,
      template_name: activeTemplate?.name,
      commit_message: commitMessage,
      parameter_values: {
        ...formData,
      },
    });
  }, [
    addCluster,
    formData,
    branchName,
    pullRequestTitle,
    commitMessage,
    activeTemplate?.name,
    infraCredential,
    pullRequestDescription,
  ]);

  const required = useMemo(() => {
    return activeTemplate?.parameters?.map(param => param.name);
  }, [activeTemplate]);

  const parameters = useMemo(() => {
    return (
      activeTemplate?.parameters?.map(param => {
        const { name, options } = param;
        if (options?.length !== 0) {
          return {
            [name]: {
              type: 'string',
              title: `${name}`,
              enum: options,
            },
          };
        } else {
          return {
            [name]: {
              type: 'string',
              title: `${name}`,
            },
          };
        }
      }) || []
    );
  }, [activeTemplate]);

  const properties = useMemo(() => {
    return Object.assign({}, ...parameters);
  }, [parameters]);

  const schema: JSONSchema7 = useMemo(() => {
    return {
      type: 'object',
      properties,
      required,
    };
  }, [properties, required]);

  // Adapted from : https://codesandbox.io/s/0y7787xp0l?file=/src/index.js:1507-1521
  const sections = useMemo(() => {
    const groups =
      activeTemplate?.objects.reduce(
        (accumulator, item, index) =>
          Object.assign(accumulator, {
            [`${index + 1}. ${item.kind}`]: item.parameters,
          }),
        {},
      ) || {};
    Object.assign(groups, { 'ui:template': 'box' });
    return [groups];
  }, [activeTemplate]);

  const uiSchema = useMemo(() => {
    return {
      'ui:groups': sections,
      'ui:template': (props: ObjectFieldTemplateProps) => (
        <Grouped.ObjectFieldTemplate {...props} />
      ),
    };
  }, [sections]);

  useEffect(() => {
    if (!activeTemplate) {
      setActiveTemplate(getTemplate(templateName));
    }
    const steps =
      activeTemplate?.objects?.map(
        (object, index) => `${index + 1}. ${object.kind}`,
      ) || [];
    setSteps(steps);
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
          <Grid container spacing={3}>
            <Grid item xs={12} md={10}>
              <div className={classes.title}>
                Create new cluster with template
              </div>
              <CredentialsWrapper>
                <div className="template-title">
                  Template: <span>{activeTemplate?.name}</span>
                </div>
                <div className="credentials">
                  <span>Infrastructure provider credentials:</span>
                  <Dropdown
                    value={infraCredential?.name}
                    disabled={loading}
                    items={credentialsItems}
                    onChange={handleSelectCredentials}
                  />
                </div>
              </CredentialsWrapper>
              <Divider
                className={
                  !isLargeScreen ? classes.divider : classes.largeDivider
                }
              />
              <Form
                className={classes.form}
                schema={schema as JSONSchema7}
                onChange={({ formData }) => setFormData(formData)}
                formData={formData}
                onSubmit={handlePreview}
                onError={() => console.log('errors')}
                uiSchema={uiSchema}
                formContext={{
                  templates: FormSteps,
                  clickedStep,
                  setActiveStep,
                }}
                {...UiTemplate}
              >
                <div className={classes.previewCTA}>
                  <Button>Preview PR</Button>
                </div>
              </Form>
              {openPreview ? (
                <>
                  {PRPreview ? (
                    <>
                      <FormStep
                        title="Preview"
                        active={activeStep === 'Preview'}
                        clicked={clickedStep === 'Preview'}
                        setActiveStep={setActiveStep}
                      >
                        <textarea
                          className={classes.textarea}
                          rows={rows}
                          value={PRPreview}
                          readOnly
                        />
                        <span>
                          You may edit these as part of the pull request with
                          your git provider.
                        </span>
                      </FormStep>
                      <FormStep
                        title="GitOps"
                        active={activeStep === 'GitOps'}
                        clicked={clickedStep === 'GitOps'}
                        setActiveStep={setActiveStep}
                      >
                        <Input
                          label="Create branch"
                          placeholder={branchName}
                          onChange={handleChangeBranchName}
                        />
                        <Input
                          label="Pull request title"
                          placeholder={pullRequestTitle}
                          onChange={handleChangePullRequestTitle}
                        />
                        <Input
                          label="Commit message"
                          placeholder={commitMessage}
                          onChange={handleChangeCommitMessage}
                        />
                        <Input
                          label="Pull request description"
                          placeholder={pullRequestDescription}
                          onChange={handleChangePRDescription}
                          multiline
                          rows={4}
                        />
                        <div
                          className={classes.createCTA}
                          onClick={handleAddCluster}
                        >
                          <Button onClick={() => setClickedStep('GitOps')}>
                            Create Pull Request
                          </Button>
                        </div>
                      </FormStep>
                      {creatingPR && <Loader />}
                    </>
                  ) : (
                    <Loader />
                  )}
                </>
              ) : null}
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
      </PageTemplate>
    );
  }, [
    clustersCount,
    activeTemplate?.name,
    classes,
    formData,
    handlePreview,
    handleAddCluster,
    schema,
    uiSchema,
    openPreview,
    PRPreview,
    rows,
    handleChangeBranchName,
    handleChangeCommitMessage,
    handleChangePullRequestTitle,
    handleChangePRDescription,
    handleSelectCredentials,
    creatingPR,
    steps,
    activeStep,
    credentialsItems,
    loading,
    infraCredential?.name,
    clickedStep,
    isLargeScreen,
    branchName,
    commitMessage,
    pullRequestTitle,
    pullRequestDescription,
  ]);
};

const AddClusterWithCredentials = () => (
  <CredentialsProvider>
    <AddCluster />
  </CredentialsProvider>
);

export default AddClusterWithCredentials;
