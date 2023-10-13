import {
  Flex,
  Icon,
  IconType,
  KubeStatusIndicator,
  Link,
  Text,
  V2Routes,
  formatURL,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';
import styled from 'styled-components';
import { PipelineTargetStatus } from '../../../api/pipelines/types.pb';
import { useListConfigContext } from '../../../contexts/ListConfig';
import { ClusterDashboardLink } from '../../Clusters/ClusterDashboardLink';
import { EnvironmentCard } from './styles';

type Props = {
  className?: string;
  target: PipelineTargetStatus;
  background: number;
};

function Target({ className, target, background }: Props) {
  const configResponse = useListConfigContext();
  const clusterName = target?.clusterRef?.name
    ? `${target?.clusterRef?.namespace || 'default'}/${
        target?.clusterRef?.name
      }`
    : configResponse?.data?.managementClusterName || 'undefined';

  //questions for Monday: clusterRefs empty? Cluster namespace? Multiple workloads?
  return (
    <EnvironmentCard
      className={className}
      background={background}
      column
      gap="8"
    >
      <Flex column>
        <Flex gap="4" align>
          <Icon type={IconType.ClustersIcon} size="medium" color="neutral20" />
          <Text>Cluster:</Text>
        </Flex>
        <ClusterDashboardLink clusterName={clusterName} />
      </Flex>
      {_.map(target.workloads, (workload, index) => {
        return (
          <Flex key={index} column gap="8">
            <Flex column>
              <Text size="medium">Namespace/Name</Text>
              <Flex gap="4" align>
                <KubeStatusIndicator
                  noText
                  conditions={workload?.conditions || []}
                />
                <Link
                  to={formatURL(V2Routes.HelmRelease, {
                    name: workload.name,
                    namespace: target.namespace,
                    clusterName: clusterName,
                  })}
                >
                  {target.namespace} / {workload.name}
                </Link>
              </Flex>
            </Flex>
            <Text bold size="small">
              LAST APPLIED VERSION: {'V' + workload.lastAppliedRevision || '-'}
            </Text>
            <Text size="small">
              SPECIFIED VERSION: {'V' + workload.version || '-'}
            </Text>
          </Flex>
        );
      })}
    </EnvironmentCard>
  );
}

export default styled(Target).attrs({ className: Target.name })``;
