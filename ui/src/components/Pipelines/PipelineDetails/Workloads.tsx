import { Grid } from '@material-ui/core';
import { Flex, formatURL, Link } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import {
  Pipeline,
  PipelineTargetStatus,
} from '../../../api/pipelines/types.pb';
import { useListConfigContext } from '../../../contexts/ListConfig';
import { ClusterDashboardLink } from '../../Clusters/ClusterDashboardLink';
import PromotePipeline from './PromotePipeline';
import {
  CardContainer,
  ClusterName,
  LastAppliedVersion,
  TargetNamespace,
  TargetWrapper,
  Title,
  usePipelineStyles,
  WorkloadWrapper,
} from './styles';
import WorkloadStatus from './WorkloadStatus';

const getTargetsCount = (targetsStatuses: PipelineTargetStatus[]) => {
  return targetsStatuses?.reduce((prev, next) => {
    return prev + (next.workloads?.length || 0);
  }, 0);
};

const TargetStatus = ({
  target,
  classes,
}: {
  target: PipelineTargetStatus;
  classes: any;
}) => {
  const configResponse = useListConfigContext();
  const clusterName = target?.clusterRef?.name
    ? `${target?.clusterRef?.namespace || 'default'}/${
        target?.clusterRef?.name
      }`
    : configResponse?.data?.managementClusterName || 'undefined';
  return (
    <>
      {target.workloads?.map((workload, wrkIndex) => (
        <CardContainer key={wrkIndex} role="targeting">
          <TargetWrapper className="workloadTarget">
            <Title>Cluster</Title>
            <ClusterName className="cluster-name">
              <ClusterDashboardLink
                clusterName={target?.clusterRef?.name || clusterName}
              />
            </ClusterName>
            <Title>Namespace</Title>
            <TargetNamespace className="workload-namespace">
              {target?.namespace}
            </TargetNamespace>
          </TargetWrapper>
          <WorkloadWrapper>
            <Flex wide between>
              <div className="automation">
                <Link
                  to={formatURL('/helm_release/details', {
                    name: workload?.name,
                    namespace: target?.namespace,
                    clusterName,
                  })}
                >
                  {workload && <WorkloadStatus workload={workload} />}
                </Link>
                <div className={`${classes.subtitle} ${classes.subtitleColor}`}>
                  <span>Specification:</span>
                  <span className={`version`}>{`v${workload?.version}`}</span>
                </div>
              </div>
              {workload?.lastAppliedRevision && (
                <LastAppliedVersion className="last-applied-version">{`v${workload?.lastAppliedRevision}`}</LastAppliedVersion>
              )}
            </Flex>
          </WorkloadWrapper>
        </CardContainer>
      ))}
    </>
  );
};

function Workloads({
  pipeline,
  className,
}: {
  pipeline: Pipeline;
  className?: string;
}) {
  const classes = usePipelineStyles();
  const environments = pipeline?.environments || [];
  const targetsStatuses = pipeline?.status?.environments || {};
  const manual = pipeline?.promotion?.manual || false;

  return (
    <Grid
      className={classes.gridWrapper + ` ${className}`}
      container
      spacing={4}
    >
      {environments.map((env, index) => {
        const status = targetsStatuses[env.name!].targetsStatuses || [];
        const promoteVersion =
          targetsStatuses[env.name!].waitingStatus?.revision || '';

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
            {true && (
              <PromotePipeline
                req={{
                  name: pipeline.name,
                  env: env.name,
                  namespace: pipeline.namespace,
                  revision: promoteVersion,
                }}
                promoteVersion={promoteVersion || ''}
              />
            )}
            {status.map((target, indx) => (
              <TargetStatus target={target} classes={classes} key={indx} />
            ))}
          </Grid>
        );
      })}
    </Grid>
  );
}

export default styled(Workloads)`
  .MuiGrid-item {
    background: ${props => props.theme.colors.neutralGray};
  }
`;
