import {
  Automation,
  Canary,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { Flex, Link, Text, formatURL } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { getKindRoute } from '../../../../utils/nav';
import { ClusterDashboardLink } from '../../../Clusters/ClusterDashboardLink';
import RowHeader from '../../../RowHeader';
import { getDeploymentStrategyIcon } from '../../ListCanaries/Table';
import Collapsible from '../../SharedComponent/Collapsible';
import DynamicTable from '../../SharedComponent/DynamicTable';

const StatusHeader = styled(Text)`
  background: ${props => props.theme.colors.neutralGray};
  padding: 16px 8px;
  margin: 16px 0px;
`;

const DetailsSection = ({
  canary,
  automation,
}: {
  canary: Canary;
  automation?: Automation;
}) => {
  const { conditions, ...restStatus } = canary?.status || { conditions: [] };
  const { lastTransitionTime, ...restConditionObj } = conditions![0] || {
    lastTransitionTime: '',
  };

  return (
    <Flex column gap="8">
      <RowHeader
        rowkey="Cluster"
        value={<ClusterDashboardLink clusterName={canary.clusterName || ''} />}
      />
      <RowHeader rowkey="Namespace" value={canary.namespace} />
      <RowHeader
        rowkey="Target"
        value={`${canary.targetReference?.kind}/${canary.targetReference?.name}`}
      />
      <RowHeader
        rowkey="Application"
        value={
          automation?.kind && automation?.name ? (
            <Link
              to={formatURL(getKindRoute(automation?.kind), {
                name: automation?.name,
                namespace: automation?.namespace,
                clusterName: canary.clusterName,
              })}
            >
              {automation?.kind}/{automation?.name}
            </Link>
          ) : (
            ''
          )
        }
      />
      <RowHeader rowkey="Deployment Strategy" value={undefined}>
        {!!canary.deploymentStrategy && (
          <span>
            {canary.deploymentStrategy}{' '}
            {getDeploymentStrategyIcon(canary.deploymentStrategy)}
          </span>
        )}
      </RowHeader>
      <RowHeader rowkey="Provider" value={canary.provider} />

      <Collapsible title="STATUS">
        <DynamicTable obj={{ ...restStatus, ...restConditionObj } || {}} />
      </Collapsible>
    </Flex>
  );
};

export default DetailsSection;
