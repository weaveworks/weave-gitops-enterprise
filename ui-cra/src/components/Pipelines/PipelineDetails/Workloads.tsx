import React from 'react';
import WorkloadStatus from './WorkloadStatus';
import { ClusterDashboardLink } from '../../Clusters/ClusterDashboardLink';
import { Flex, formatURL, Link } from '@weaveworks/weave-gitops';
import {
  Pipeline,
  PipelineTargetStatus,
} from '../../../api/pipelines/types.pb';
import useConfig from '../../../hooks/config';

import { Grid } from '@material-ui/core';
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

const getTargetsCount = (targetsStatuses: PipelineTargetStatus[]) => {
  return targetsStatuses?.reduce((prev, next) => {
    return prev + (next.workloads?.length || 0);
  }, 0);
};

function Workloads({ pipeline }: { pipeline: Pipeline }) {
  const classes = usePipelineStyles();
  const configResponse = useConfig();
  const environments = pipeline?.environments || [];
  const targetsStatuses = pipeline?.status?.environments || {};

  return (
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
            {status.map(target => {
              const clusterName = target?.clusterRef?.name
                ? `${target?.namespace}/${target?.clusterRef?.name}`
                : configResponse?.data?.managementClusterName || 'undefined';

              return target?.workloads?.map((workload, wrkIndex) => (
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
                    </div>
                    <Flex wide between>
                      <div
                        style={{ alignSelf: 'flex-end' }}
                        className={`${classes.subtitle} ${classes.subtitleColor}`}
                      >
                        <span>Specification:</span>
                        <span className={`version`}>
                          {`v${workload?.version}`}
                        </span>
                      </div>
                      {workload?.lastAppliedRevision && (
                        <LastAppliedVersion className="last-applied-version">{`v${workload?.lastAppliedRevision}`}</LastAppliedVersion>
                      )}
                    </Flex>
                  </WorkloadWrapper>
                </CardContainer>
              ));
            })}
          </Grid>
        );
      })}
    </Grid>
  );
}

export default Workloads;
