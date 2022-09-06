import { createStyles, Grid, makeStyles } from '@material-ui/core';
import { ThemeProvider } from '@material-ui/core/styles';
import { theme } from '@weaveworks/weave-gitops';
import { useCountPipelines, useGetPipeline } from '../../../contexts/Pipelines';
import { localEEMuiTheme } from '../../../muiTheme';
import { useApplicationsCount } from '../../Applications/utils';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ChipWrapper } from '../../Policies/PolicyStyles';

const { medium, xs, xxs } = theme.spacing;
const { small, large } = theme.fontSizes;
const { neutral30 } = theme.colors;
const useStyles = makeStyles(() =>
  createStyles({
    gridWrapper: {
      boxShadow: '0px 1px 2px rgba(26, 26, 26, 0.24)',
      margin: medium,
      borderRadius: '10px',
    },
    title: {
      fontSize: large,
      fontWeight: 600,
    },
    subtitle: {
      fontSize: small,
      fontWeight: 600,
      color: neutral30,
      paddingTop: xs,
    },
    listVersion: {
      display: 'inline-flex',
      marginTop: xxs,
    },
  }),
);

const PipelineDetails = ({
  name,
  namespace,
}: {
  name: string;
  namespace: string;
}) => {
  const applicationsCount = useApplicationsCount();
  const pipelinesCount = useCountPipelines();
  const { isLoading, error, data } = useGetPipeline({ name, namespace });
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
              label: name,
            },
          ]}
        />
        <ContentWrapper loading={isLoading} errorMessage={error?.message}>
          {data?.pipeline && (
            <Grid container spacing={3}>
              {Object.entries(data.pipeline.status?.environments || {}).map(
                ([key, value], index) => (
                  <Grid item xs key={index} className={classes.gridWrapper}>
                    <div className={classes.title}>{key}</div>
                    <div>
                      <div className={classes.subtitle}>Target</div>
                      <div>
                        {value.clusterRef?.name}/{value.namespace}
                      </div>
                    </div>

                    <div>
                      <div className={classes.subtitle}>Name</div>
                      <span>{value.workloads![0].name}</span>
                    </div>
                    <div>
                      <div className={classes.subtitle}>Version(s)</div>
                      <div className={classes.listVersion}>
                        {value.workloads?.map((wrk, indx) => (
                          <div key={indx}>
                            <ChipWrapper>v{wrk.version}</ChipWrapper>
                          </div>
                        ))}
                      </div>
                    </div>
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
