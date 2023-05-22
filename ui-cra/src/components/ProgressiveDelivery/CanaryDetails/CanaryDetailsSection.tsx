import { RouterTab, SubRouterTabs, YamlView } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { useCanaryStyle } from '../CanaryStyles';

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
  const classes = useCanaryStyle();
  const path = Routes.CanaryDetails;
  return (
    <>
      <TitleWrapper>{canary.name}</TitleWrapper>
      <div className={classes.statusWrapper}>
        <CanaryStatus
          status={canary.status?.phase || '--'}
          value={getProgressValue(
            canary.deploymentStrategy || '',
            canary.status,
            canary.analysis,
          )}
        />
        <p className={classes.statusMessage}>
          {canary.status?.conditions![0].message || '--'}
        </p>
      </div>

      <SubRouterTabs>
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
    </>
  );
}

export default CanaryDetailsSection;
