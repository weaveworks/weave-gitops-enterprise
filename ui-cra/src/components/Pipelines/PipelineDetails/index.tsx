import { createStyles, Grid, makeStyles } from '@material-ui/core';
import { ThemeProvider } from '@material-ui/core/styles';
import { theme, Link } from '@weaveworks/weave-gitops';
import { useCountPipelines, useGetPipeline } from '../../../contexts/Pipelines';
import { localEEMuiTheme } from '../../../muiTheme';
import { useApplicationsCount } from '../../Applications/utils';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';

const { medium, xs, xxs, large } = theme.spacing;
const { small } = theme.fontSizes;
const { white, neutral10 } = theme.colors;
const useStyles = makeStyles(() =>
  createStyles({
    gridContainer: {
      backgroundColor: neutral10,
      margin: `0 ${small}`,
      padding: medium,
      borderRadius: '10px',
    },
    gridWrapper: {
      maxWidth: 'calc(100vw - 300px)',
      display: 'flex',
      flexWrap: 'nowrap',
      overflow: 'auto',
      paddingBottom: '8px',

      margin: `${medium} 0 0 0`,
    },
    title: {
      fontSize: `calc(${small} + ${small})`,
      fontWeight: 600,
      textTransform: 'capitalize',
    },
    cardContainer: {
      background: white,
      padding: medium,
      marginBottom: xs,
      boxShadow: '0px 0px 1px rgba(26, 32, 36, 0.32)',
      borderRadius: '10px',
    },
    subtitle: {
      fontSize: small,
      fontWeight: 400,
      marginTop: xxs,
    },
    cardHeader: {
      marginBottom: small,
    },
    target: {
      fontSize: '20px',
      marginBottom: '12px',
      textOverflow: 'ellipsis',
      whiteSpace: 'nowrap',
      overflow: 'hidden',
      width: `calc(250px - ${large})`,
    },
  }),
);

const PipelineDetails = ({
  name,
  namespace,
  pipelineName,
}: {
  name: string;
  namespace: string;
  pipelineName: string;
}) => {
  const applicationsCount = useApplicationsCount();
  const pipelinesCount = useCountPipelines();
  const { isLoading, error, data } = useGetPipeline({
    name,
    namespace,
  });
  const classes = useStyles();
  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Pipeline Details">
        <SectionHeader
          className="count-header"
          path={[
            {
              label: 'Applications',
              url: '/applications',
              count: applicationsCount,
            },
            {
              label: 'Pipelines',
              url: '/applications/pipelines',
              count: pipelinesCount,
            },
            {
              label: pipelineName,
            },
          ]}
        />
        <ContentWrapper loading={isLoading} errorMessage={error?.message}>
          {data?.pipeline && (
            <Grid className={classes.gridWrapper} container spacing={8}>
              {Object.entries(data.pipeline.status?.environments || {}).map(
                ([key, value], index) => (
                  <Grid item xs key={index} className={classes.gridContainer}>
                    <div className={classes.cardHeader}>
                      <div className={classes.title}>{key}</div>
                      <div className={classes.subtitle}>
                        {value.length} Targets
                      </div>
                    </div>
                    {value.map((target, indx) => (
                      <div className={classes.cardContainer} key={indx}>
                        <div
                          className={classes.target}
                          title={`${target.clusterRef?.name}/${target.namespace}`}
                        >
                          {target.clusterRef?.name}/{target.namespace}
                        </div>
                        <div>
                          <Link
                            to={`/helm_release/details?clusterName=${target.clusterRef?.name}&name=${name}&namespace=${namespace}`}
                          >
                            {target.workloads![0].name}
                          </Link>
                          <div className={classes.subtitle}>Version</div>
                          <div className={classes.subtitle}>{`v${
                            target.workloads![0].version
                          }`}</div>
                        </div>
                      </div>
                    ))}
                  </Grid>
                ),
              )}
            </Grid>
          )}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PipelineDetails;
