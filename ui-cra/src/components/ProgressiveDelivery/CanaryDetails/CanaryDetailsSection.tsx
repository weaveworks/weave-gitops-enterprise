import { Table, TableBody } from '@material-ui/core';
import { RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { useCanaryStyle } from '../CanaryStyles';
import CanaryRowHeader, {
  KeyValueRow,
} from '../SharedComponent/CanaryRowHeader';
import CanaryStatus from '../SharedComponent/CanaryStatus';

import { ExpandLess, ExpandMore } from '@material-ui/icons';
import {
  Automation,
  Canary,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { useState } from 'react';
import YamlView from '../../YamlView';
import ListEvents from '../Events/ListEvents';
import {
  getDeploymentStrategyIcon,
  getProgressValue,
} from '../ListCanaries/Table';
import ListManagedObjects from './ListManagedObjects';
import {CanaryMetricsTable} from "./Analysis/CanaryMetricsTable";

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
  const [open, setOpen] = useState(true);

  const { conditions, ...restStatus } = canary?.status || { conditions: [] };
  const { lastTransitionTime, ...restConditionObj } = conditions![0] || {
    lastTransitionTime: '',
  };

  const toggleCollapse = () => {
    setOpen(!open);
  };

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
            <CanaryRowHeader rowkey="Cluster" value={canary.clusterName} />
            <CanaryRowHeader rowkey="Namespace" value={canary.namespace} />
            <CanaryRowHeader
              rowkey="Target"
              value={`${canary.targetReference?.kind}/${canary.targetReference?.name}`}
            />
            <CanaryRowHeader
              rowkey="Application"
              value={
                automation?.kind && automation?.name
                  ? `${automation?.kind}/${automation?.name}`
                  : '--'
              }
            />
            <CanaryRowHeader
              rowkey="Deployment Strategy"
              value={canary.deploymentStrategy}
            >
              <span className={classes.straegyIcon}>
                {getDeploymentStrategyIcon(canary.deploymentStrategy || '')}
              </span>
            </CanaryRowHeader>
            <CanaryRowHeader rowkey="Provider" value={canary.provider} />

            <div
              className={`${classes.sectionHeaderWrapper} ${classes.cardTitle}`}
            >
              Status
            </div>

            <Table size="small">
              <TableBody>
                {Object.entries(restStatus || {}).map((entry, index) => (
                  <KeyValueRow entryObj={entry} key={index} />
                ))}
              </TableBody>
            </Table>

            <div
              className={` ${classes.cardTitle} ${classes.expandableCondition}`}
              onClick={toggleCollapse}
            >
              {!open ? <ExpandLess /> : <ExpandMore />}
              <span className={classes.expandableSpacing}> Conditions</span>
            </div>
            <Table
              size="small"
              className={open ? classes.fadeIn : classes.fadeOut}
            >
              <TableBody>
                {Object.entries(restConditionObj).map((entry, index) => (
                  <KeyValueRow entryObj={entry} key={index}></KeyValueRow>
                ))}
              </TableBody>
            </Table>
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
            <CanaryMetricsTable metrics={canary.analysis?.metrics || []}></CanaryMetricsTable>
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
