import { createStyles, Grid, makeStyles } from '@material-ui/core';
import styled from 'styled-components';
import { formatURL, Link, theme } from '@weaveworks/weave-gitops';
import { PipelineTargetStatus } from '../../../api/pipelines/types.pb';
import { useGetPipeline } from '../../../contexts/Pipelines';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { Routes } from '../../../utils/nav';

const { medium, xs, xxs, large } = theme.spacing;
const { small } = theme.fontSizes;
const { white, neutral10, neutral20, neutral30, black } = theme.colors;

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
    subtitle: {
      fontSize: small,
      fontWeight: 400,
      marginTop: xxs,
    },
    mbSmall: {
      marginBottom: small,
    },
    subtitleColor: {
      color: neutral30,
    },
  }),
);
const TargetWrapper = styled.div`
  font-size: ${theme.fontSizes.large};
  margin-bottom: ${small};
  text-overflow: ellipsis;
  white-space: nowrap;
  overflow: hidden;
  width: calc(250px - ${large});
`;
const CardContainer = styled.div`
  background: ${white};
  padding: ${small};
  margin-bottom: ${xs};
  box-shadow: 0px 0px 1px rgba(26, 32, 36, 0.32);
  border-radius: 10px;
  font-weight: 600;
`;
const Title = styled.div`
  font-size: ${theme.fontSizes.medium};
  color: ${black};
  font-weight: 400;
`;
const ClusterName = styled.div`
  margin-bottom: ${small};
  line-height: 24px;
`;
const TargetNamespace = styled.div`
  font-size: ${theme.fontSizes.medium};
`;
const WorkloadWrapper = styled.div`
  position: relative;
  .version {
    margin-left: ${xxs};
  }
  .workloadName {
    margin-bottom: ${small};
  }
`;
const LastAppliedVersion = styled.span`
  color: ${neutral30};
  position: absolute;
  right: 0;
  bottom: auto;
  font-size: ${theme.fontSizes.medium};
  border: 1px solid ${neutral20};
  padding: 14px 6px;
  border-radius: 50%;
`;
const getTargetsCount = (targetsStatuses: PipelineTargetStatus[]) => {
  return targetsStatuses?.reduce((prev, next) => {
    return prev + (next.workloads?.length || 0);
  }, 0);
};

interface Props {
  name: string;
  namespace: string;
}

const PipelineDetails = ({ name, namespace }: Props) => {
  const { isLoading, error, data } = useGetPipeline({
    name,
    namespace,
  });

  const environments = data?.pipeline?.environments || [];
  const targetsStatuses = data?.pipeline?.status?.environments || {};
  const classes = useStyles();
  return (
    <PageTemplate
      documentTitle="Pipeline Details"
      path={[
        {
          label: 'Applications',
          url: Routes.Applications,
        },
        {
          label: 'Pipelines',
          url: Routes.Pipelines,
        },
        {
          label: name,
        },
      ]}
    >
      <ContentWrapper loading={isLoading} errorMessage={error?.message}>
        <Grid className={classes.gridWrapper} container spacing={4}>
          {environments.map((env, index) => {
            const status = targetsStatuses[env.name!].targetsStatuses || [];
            return (
              <Grid
                item
                xs
                key={index}
                className={classes.gridContainer}
                id={env.name}
              >
                <div className={classes.mbSmall}>
                  <div className={classes.title}>{env.name}</div>
                  <div className={classes.subtitle}>
                    {getTargetsCount(status || [])} Targets
                  </div>
                </div>
                {status.map(target =>
                  target?.workloads?.map((workload, wrkIndex) => (
                    <CardContainer key={wrkIndex} role="targeting">
                      <TargetWrapper className="workloadTarget">
                        {target?.clusterRef?.name && (
                          <>
                            <Title>Cluster</Title>
                            <ClusterName className="cluster-name">
                              {target?.clusterRef?.name}
                            </ClusterName>
                          </>
                        )}

                        <Title>Namespace</Title>
                        <TargetNamespace className="workload-namespace">
                          {target?.namespace}
                        </TargetNamespace>
                      </TargetWrapper>
                      <WorkloadWrapper>
                        <div className={`workloadName`}>
                          {target?.clusterRef?.namespace ? (
                            <Link
                              to={formatURL('/helm_release/details', {
                                name: workload?.name,
                                namespace: target?.namespace,
                                clusterName: `${target?.clusterRef?.namespace}/${target?.clusterRef?.name}`,
                              })}
                            >
                              {workload?.name}
                            </Link>
                          ) : (
                            <div className="workload-name">
                              {workload?.name}
                            </div>
                          )}
                          {workload?.lastAppliedRevision && (
                            <LastAppliedVersion className="last-applied-version">{`v${workload?.lastAppliedRevision}`}</LastAppliedVersion>
                          )}
                        </div>
                        <div
                          className={`${classes.subtitle} ${classes.subtitleColor}`}
                        >
                          <span>Specification:</span>
                          <span className={`version`}>
                            {`v${workload?.version}`}
                          </span>
                        </div>
                      </WorkloadWrapper>
                    </CardContainer>
                  )),
                )}
              </Grid>
            );
          })}
        </Grid>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default PipelineDetails;
