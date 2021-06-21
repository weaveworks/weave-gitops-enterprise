import React, {
  ChangeEvent,
  FC,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react';
import styled from 'styled-components';
import useTemplates from '../../../contexts/Templates';
import useClusters from '../../../contexts/Clusters';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import Divider from '@material-ui/core/Divider';
import { useHistory, useParams } from 'react-router-dom';
import Form from '@rjsf/material-ui';
import { JSONSchema7 } from 'json-schema';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { Button } from 'weaveworks-ui-components';
import { ISubmitEvent } from '@rjsf/core';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import Grid from '@material-ui/core/Grid';
import { NavItem } from '../../Navigation';
import { Input } from '../../../utils/form';

const StepNavItem = styled(NavItem)`
  font-size: ${weaveTheme.fontSizes.normal};
`;

const base = weaveTheme.spacing.base;
const xxs = weaveTheme.spacing.xxs;
const xs = weaveTheme.spacing.xs;
const small = weaveTheme.spacing.small;

const useStyles = makeStyles(theme =>
  createStyles({
    form: {
      paddingTop: base,
    },
    title: {
      color: `${weaveTheme.colors.gray600}`,
    },
    sectionTitle: {
      paddingTop: base,
      paddingBottom: base,
    },
    divider: {
      marginTop: xxs,
      marginBottom: xxs,
    },
    textarea: {
      width: '100%',
      padding: xs,
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
    main: {
      order: 1,
      [theme.breakpoints.down('sm')]: {
        order: 2,
      },
    },
    steps: {
      display: 'flex',
      flexDirection: 'column',
      order: 2,
      [theme.breakpoints.down('sm')]: {
        order: 1,
        flexDirection: 'row',
      },
      marginTop: base,
      paddingRight: xxs,
    },
  }),
);

const AddCluster: FC = () => {
  const classes = useStyles();
  const {
    activeTemplate,
    setActiveTemplate,
    renderTemplate,
    PRPreview,
    addCluster,
    getTemplate,
  } = useTemplates();
  const { count } = useClusters();
  const [formData, setFormData] = useState({});
  const [openPreview, setOpenPreview] = useState(false);
  const [branchName, setBranchName] = useState<string>('default');
  const [pullRequestTitle, setPullRequestTitle] = useState<string>('default');
  const [commitMessage, setCommitMessage] = useState<string>('default');
  const history = useHistory();
  const rows = (PRPreview?.split('\n').length || 0) - 1;
  const { templateName } = useParams<{ templateName: string }>();

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
    addCluster({ ...formData, branchName, pullRequestTitle, commitMessage });
    history.push('/clusters');
  }, [
    addCluster,
    formData,
    history,
    branchName,
    pullRequestTitle,
    commitMessage,
  ]);

  const required = useMemo(() => {
    return activeTemplate?.parameters?.map(param => param.name);
  }, [activeTemplate]);

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
      title: 'Cluster',
      type: 'object',
      required,
      properties,
    };
  }, [properties, required]);

  useEffect(() => {
    if (!activeTemplate) {
      setActiveTemplate(getTemplate(templateName));
    }
  }, [activeTemplate, getTemplate, setActiveTemplate, templateName]);

  return useMemo(() => {
    return (
      <PageTemplate documentTitle="WeGo · Create new cluster">
        <SectionHeader
          path={[
            { label: 'Clusters', url: '/', count },
            { label: 'Create new cluster' },
          ]}
        />
        <ContentWrapper>
          <Grid container spacing={3}>
            <Grid className={classes.main} item xs={12} md={10}>
              <h3 className={classes.title}>
                Create new cluster with template
              </h3>
              Template: {activeTemplate?.name}
              <Form
                className={classes.form}
                schema={schema as JSONSchema7}
                onChange={() => console.log('changed')}
                formData={formData}
                onSubmit={handlePreview}
                onError={() => console.log('errors')}
              >
                <div className={classes.previewCTA}>
                  <Button>Preview PR</Button>
                </div>
              </Form>
              {openPreview ? (
                <div>
                  <Divider className={classes.divider} />
                  <div className={classes.sectionTitle}>
                    <span>Preview & Commit</span>
                  </div>
                  <textarea
                    className={classes.textarea}
                    rows={rows}
                    value={PRPreview || 'No preview available'}
                    readOnly
                  />
                  You may edit these as part of the pull request with your git
                  provider.
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
                  <div className={classes.createCTA} onClick={handleAddCluster}>
                    <Button>Create Pull Request on GitHub</Button>
                  </div>
                </div>
              ) : null}
            </Grid>
            <Grid
              className={classes.steps}
              item
              md={2}
              style={{ visibility: 'hidden' }}
            >
              {['step1', 'step2', 'step3', 'step4', 'step5'].map(step => (
                <StepNavItem to={`#${step}`}>{step}</StepNavItem>
              ))}
            </Grid>
          </Grid>
        </ContentWrapper>
      </PageTemplate>
    );
  }, [
    count,
    activeTemplate?.name,
    classes,
    formData,
    handlePreview,
    handleAddCluster,
    schema,
    openPreview,
    PRPreview,
    rows,
    handleChangeBranchName,
    handleChangeCommitMessage,
    handleChangePullRequestTitle,
  ]);
};

export default AddCluster;
