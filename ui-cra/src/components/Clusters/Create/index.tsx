import React, {
  ChangeEvent,
  FC,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react';
import useTemplates from '../../../contexts/Templates';
import useClusters from '../../../contexts/Clusters';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { useParams } from 'react-router-dom';
import Form from '@rjsf/material-ui';
import { JSONSchema7 } from 'json-schema';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { Button } from 'weaveworks-ui-components';
import { ISubmitEvent, ObjectFieldTemplateProps } from '@rjsf/core';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import Grid from '@material-ui/core/Grid';
import { Input } from '../../../utils/form';
import { Loader } from '../../Loader';
import Divider from '@material-ui/core/Divider';
import { useHistory } from 'react-router-dom';
import * as Grouped from './Form/GroupedSchema';
import * as UiTemplate from './Form/UITemplate';
import FormSteps from './Form/Steps';
import FormStepsNavigation from './Form/StepsNavigation';

const medium = weaveTheme.spacing.medium;
const base = weaveTheme.spacing.base;
const xxs = weaveTheme.spacing.xxs;
const xs = weaveTheme.spacing.xs;
const small = weaveTheme.spacing.small;

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
    sectionTitle: {
      fontSize: weaveTheme.fontSizes.large,
      paddingTop: base,
      paddingBottom: base,
    },
    divider: {
      marginTop: medium,
      marginBottom: base,
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
    grid: {
      [theme.breakpoints.down('md')]: {
        justifyContent: 'flex-end',
      },
    },
    main: {},
    steps: {
      display: 'flex',
      flexDirection: 'column',
      [theme.breakpoints.down('sm')]: {
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
    PRurl,
    setPRurl,
    creatingPR,
    addCluster,
  } = useTemplates();
  const clustersCount = useClusters().count;
  const [formData, setFormData] = useState({});
  const [steps, setSteps] = useState<string[]>([]);
  const [openPreview, setOpenPreview] = useState(false);
  const [branchName, setBranchName] = useState<string>('default');
  const [pullRequestTitle, setPullRequestTitle] = useState<string>('default');
  const [commitMessage, setCommitMessage] = useState<string>('default');
  const rows = (PRPreview?.split('\n').length || 0) - 1;
  const { templateName } = useParams<{ templateName: string }>();
  const history = useHistory();
  const [activeStep, setActiveStep] = useState<string>('');

  const handlePreview = useCallback(
    (event: ISubmitEvent<any>) => {
      setFormData(event.formData);
      setOpenPreview(true);
      renderTemplate(event.formData);
    },
    [setOpenPreview, setFormData, renderTemplate],
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

  const handleAddCluster = useCallback(() => {
    addCluster({
      head_branch: branchName,
      title: pullRequestTitle,
      description: 'This PR creates a new Kubernetes cluster',
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
  ]);

  const parameters = useMemo(() => {
    return (
      activeTemplate?.parameters?.map(param => {
        const name = param.name;
        return {
          [name]: { type: 'string', title: `${name}` },
        };
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
    };
  }, [properties]);

  // Adapted from : https://codesandbox.io/s/0y7787xp0l?file=/src/index.js:1507-1521
  const sections = useMemo(() => {
    const groups =
      activeTemplate?.objects.reduce(
        (accumulator, item) =>
          Object.assign(accumulator, {
            [item.kind]: item.parameters,
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
    const steps = activeTemplate?.objects?.map(object => object.kind) || [];
    setSteps(steps);
    return history.listen(() => {
      setPRurl(null);
      setActiveTemplate(null);
    });
  }, [
    activeTemplate,
    getTemplate,
    setActiveTemplate,
    templateName,
    setPRurl,
    history,
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
          <Grid className={classes.grid} container spacing={3}>
            <Grid className={classes.main} item xs={12} md={9}>
              <div className={classes.title}>
                Create new cluster with template
              </div>
              Template: {activeTemplate?.name}
              <Form
                className={classes.form}
                schema={schema as JSONSchema7}
                onChange={() => console.log('changed')}
                formData={formData}
                onSubmit={handlePreview}
                onError={() => console.log('errors')}
                uiSchema={uiSchema}
                formContext={{
                  templates: FormSteps,
                  activeStep,
                }}
                {...UiTemplate}
              >
                <div className={classes.previewCTA}>
                  <Button>Preview PR</Button>
                </div>
              </Form>
              {openPreview ? (
                <>
                  <div className={classes.sectionTitle}>
                    <span>Preview & Commit</span>
                  </div>
                  {PRPreview ? (
                    <>
                      <textarea
                        className={classes.textarea}
                        rows={rows}
                        value={PRPreview}
                        readOnly
                      />
                      <span>
                        You may edit these as part of the pull request with your
                        git provider.
                      </span>
                      <>
                        <Divider className={classes.divider} />
                        <div className={classes.sectionTitle}>
                          <span>GitOps</span>
                        </div>
                        <Input
                          label="Create branch"
                          onChange={handleChangeBranchName}
                        />
                        <Input
                          label="Title pull request"
                          onChange={handleChangePullRequestTitle}
                        />
                        <Input
                          label="Commit message"
                          onChange={handleChangeCommitMessage}
                        />
                        <div
                          className={classes.createCTA}
                          onClick={handleAddCluster}
                        >
                          <Button>Create Pull Request on GitHub</Button>
                        </div>
                        {creatingPR ? (
                          <>
                            <Divider className={classes.divider} />
                            <Loader />
                          </>
                        ) : null}
                        {PRurl && !creatingPR ? (
                          <>
                            <Divider className={classes.divider} />
                            <span>
                              You can access your newly created PR&nbsp;
                              <a href={PRurl} target="_blank" rel="noreferrer">
                                here.
                              </a>
                            </span>
                          </>
                        ) : null}
                      </>
                    </>
                  ) : (
                    <Loader />
                  )}
                </>
              ) : null}
            </Grid>
            <Grid className={classes.steps} item md={3}>
              <FormStepsNavigation
                steps={steps}
                activeStep={activeStep}
                setActiveStep={setActiveStep}
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
    PRurl,
    rows,
    handleChangeBranchName,
    handleChangeCommitMessage,
    handleChangePullRequestTitle,
    creatingPR,
    steps,
    activeStep,
  ]);
};

export default AddCluster;
