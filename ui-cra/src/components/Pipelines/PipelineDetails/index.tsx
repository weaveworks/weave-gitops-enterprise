import { createStyles, Grid, makeStyles } from '@material-ui/core';
import { theme } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { PipelineTargetStatus } from '../../../api/pipelines/types.pb';
import { useCountPipelines, useGetPipeline } from '../../../contexts/Pipelines';
import { useApplicationsCount } from '../../Applications/utils';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';

const { medium, xs, xxs, large } = theme.spacing;
const { small } = theme.fontSizes;
const { white, neutral10, neutral30 } = theme.colors;

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
    subtitleColor: {
      color: neutral30,
    },
  }),
);
const getTargetsCount = (targetsStatuses: PipelineTargetStatus[]) => {
  return targetsStatuses?.reduce((prev, next) => {
    return prev + (next.workloads?.length || 0);
  }, 0);
};

interface Props {
  name: string;
  namespace: string;
  pipelineName: string;
}

const PipelineDetails = ({ name, namespace, pipelineName }: Props) => {
  const applicationsCount = useApplicationsCount();
  const pipelinesCount = useCountPipelines();
  const { isLoading, error, data } = useGetPipeline({
    name,
    namespace,
  });

  const environments = data?.pipeline?.status?.environments || {};

  const classes = useStyles();
  return (
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
        <Grid className={classes.gridWrapper} container spacing={8}>
          {Object.entries(environments).map(([envName, envStatus], index) => (
            <Grid item xs key={envName} className={classes.gridContainer}>
              <div className={classes.cardHeader}>
                <div className={classes.title}>{envName}</div>
                <div className={classes.subtitle}>
                  {getTargetsCount(envStatus.targetsStatuses || [])} Targets
                </div>
              </div>
              {_.map(envStatus.targetsStatuses, target =>
                _.map(target.workloads, (workload, wrkIndex) => (
                  <div className={classes.cardContainer} key={wrkIndex}>
                    <div className={classes.target}>
                      {target?.clusterRef?.name
                        ? `${target?.clusterRef?.name}/${target?.namespace}`
                        : target?.namespace}
                    </div>
                    <div>
                      {workload?.name}
                      {/* <Link
                     to={`/helm_release/details?clusterName=${target?.clusterRef?.name}&name=${name}&namespace=${namespace}`}
                   >
                     {workload?.name}
                   </Link> */}
                      <div
                        className={`${classes.subtitle} ${classes.subtitleColor}`}
                      >
                        Version
                      </div>
                      <div
                        className={classes.subtitle}
                      >{`v${workload?.version}`}</div>
                    </div>
                  </div>
                )),
              )}
            </Grid>
          ))}
        </Grid>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default PipelineDetails;
