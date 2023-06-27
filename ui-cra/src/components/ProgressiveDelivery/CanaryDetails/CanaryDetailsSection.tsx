import {
  Flex,
  RouterTab,
  SubRouterTabs,
  Text,
  YamlView,
} from '@weaveworks/weave-gitops';
import styled from 'styled-components';

import {
  Automation,
  Canary,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { Routes } from '../../../utils/nav';
import { getProgressValue } from '../ListCanaries/Table';
import CanaryStatus from '../SharedComponent/CanaryStatus';
import { CanaryMetricsTable } from './Analysis/CanaryMetricsTable';
import DetailsSection from './Details/DetailsSection';
import ListEvents from './Events/ListEvents';
import ListManagedObjects from './ManagedObjects/ListManagedObjects';

const TitleWrapper = styled.h2`
  margin: 0px;
`;
const CanaryDetailsWrapper = styled.div`
  width: 100%;
`;

function CanaryDetailsSection({
  canary,
  automation,
}: {
  canary: Canary;
  automation?: Automation;
}) {
  const path = Routes.CanaryDetails;
  return (
    <Flex column gap="16">
      <TitleWrapper>{canary.name}</TitleWrapper>
      <Flex gap="8" align start>
        <CanaryStatus
          status={canary.status?.phase || '--'}
          value={getProgressValue(
            canary.deploymentStrategy || '',
            canary.status,
            canary.analysis,
          )}
        />
        <Text color="neutral30">
          {canary.status?.conditions![0].message || '--'}
        </Text>
      </Flex>
      <SubRouterTabs rootPath={`${path}/details`}>
        <RouterTab name="Details" path={`${path}/details`}>
          <CanaryDetailsWrapper>
            <DetailsSection canary={canary} automation={automation} />
          </CanaryDetailsWrapper>
        </RouterTab>

        <RouterTab name="Objects" path={`${path}/objects`}>
          <CanaryDetailsWrapper>
            <ListManagedObjects
              clusterName={canary.clusterName || ''}
              name={canary.name || ''}
              namespace={canary.namespace || ''}
            />
          </CanaryDetailsWrapper>
        </RouterTab>
        <RouterTab name="Events" path={`${path}/events`}>
          <CanaryDetailsWrapper>
            <ListEvents
              clusterName={canary?.clusterName}
              involvedObject={{
                kind: 'Canary',
                name: canary.name,
                namespace: canary?.namespace,
              }}
            />
          </CanaryDetailsWrapper>
        </RouterTab>
        <RouterTab name="Analysis" path={`${path}/analysis`}>
          <CanaryDetailsWrapper>
            <CanaryMetricsTable
              metrics={canary.analysis?.metrics || []}
            ></CanaryMetricsTable>
          </CanaryDetailsWrapper>
        </RouterTab>
        <RouterTab name="yaml" path={`${path}/yaml`}>
          <CanaryDetailsWrapper>
            <YamlView
              yaml={canary.yaml || ''}
              object={{
                kind: 'Canary',
                name: canary?.name,
                namespace: canary?.namespace,
              }}
            />
          </CanaryDetailsWrapper>
        </RouterTab>
      </SubRouterTabs>
    </Flex>
  );
}

export default CanaryDetailsSection;
