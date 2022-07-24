import { RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { useCanaryStyle } from '../CanaryStyles';

import CanaryStatus from '../SharedComponent/CanaryStatus';
import {
  Automation,
  Canary,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import YamlView from '../../YamlView';
import ListEvents from '../Events/ListEvents';
import { getProgressValue } from '../ListCanaries/Table';
import ListManagedObjects from './ManagedObjects/ListManagedObjects';
import { CanaryMetricsTable } from './Analysis/CanaryMetricsTable';
import DetailsSection from './Details/DetailsSection';

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
  const path = `/applications/delivery/${canary.targetDeployment?.uid}`;

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
            <YamlView yaml={canary.yaml || ''} kind="Canary" object={canary} />
          </CanaryDetailsWrapper>
        </RouterTab>
      </SubRouterTabs>
    </>
  );
}

export default CanaryDetailsSection;
