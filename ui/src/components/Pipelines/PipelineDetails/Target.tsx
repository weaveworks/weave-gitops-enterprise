import {
  Flex,
  Icon,
  IconType,
  KubeStatusIndicator,
  Link,
  Text,
  V2Routes,
  computeReady,
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
  const clusterName = target.clusterRef?.name
    ? `${target.clusterRef?.namespace || 'default'}/${target.clusterRef?.name}`
    : configResponse?.data?.managementClusterName || null;

  return (
    <EnvironmentCard
      className={className}
      background={background}
      column
      gap="8"
    >
      <Flex column>
        <Flex gap="4" align>
          <Icon type={IconType.ClustersIcon} size="medium" color="black" />
          <Text>Cluster:</Text>
        </Flex>
        <ClusterDashboardLink clusterName={clusterName} />
      </Flex>
      {_.map(target.workloads, (workload, index) => {
        const { lastAppliedRevision, version } = workload;
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
            <Text
              bold
              size="small"
              color={
                computeReady(workload?.conditions || []) === 'Ready'
                  ? 'successOriginal'
                  : 'alertOriginal'
              }
            >
              LAST APPLIED VERSION:{' '}
              {lastAppliedRevision ? 'v' + lastAppliedRevision : '-'}
            </Text>
            <Text size="small">
              SPECIFIED VERSION: {version ? 'v' + version : '-'}
            </Text>
          </Flex>
        );
      })}
    </EnvironmentCard>
  );
}

export default styled(Target).attrs({ className: Target.name })``;
