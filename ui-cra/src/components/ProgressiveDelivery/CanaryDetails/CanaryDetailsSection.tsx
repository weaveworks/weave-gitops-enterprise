import CanaryRowHeader from '../SharedComponent/CanaryRowHeader';
import CanaryStatus from '../SharedComponent/CanaryStatus';
import { useCanaryStyle } from '../CanaryStyles';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
} from '@material-ui/core';
import styled from 'styled-components';

import { Automation, Canary } from '../../../cluster-services/types.pb';
import { getDeploymentStrategyIcon } from '../ListCanaries/Table';
const TitleWrapper = styled.h2`
  margin: 0px;
`;

function CanaryDetailsSection({
  canary,
  automation,
}: {
  canary: Canary;
  automation: Automation;
}) {
  const classes = useCanaryStyle();

  return (
    <>
      <TitleWrapper>{canary.name}</TitleWrapper>
      <div className={classes.statusWrapper}>
        <CanaryStatus
          status={canary.status?.phase || '--'}
          canaryWeight={canary.status?.canaryWeight || 0}
        />
        <p className={classes.statusMessage}>
          {canary.status?.conditions![0].message || '--'}
        </p>
      </div>
      <CanaryRowHeader rowkey="Cluster" value={canary.clusterName} />
      <CanaryRowHeader rowkey="Namespace" value={canary.namespace} />
      <CanaryRowHeader
        rowkey="Target"
        value={`${automation.kind}/${automation.name}`}
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
      <CanaryRowHeader
        rowkey="Last Transition Time"
        value={canary.status?.lastTransitionTime}
      />
      <CanaryRowHeader
        rowkey="Last Updated Time"
        value={canary.status?.conditions![0].lastUpdateTime}
      />

      <div className={`${classes.sectionHeaderWrapper} ${classes.cardTitle}`}>
        Status
      </div>

      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell align="left">
              <span>Status Conditions</span>
            </TableCell>
            <TableCell align="left">
              <span>Value</span>
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          <TableRow>
            <TableCell>Canary Weight</TableCell>
            <TableCell>{canary.status?.canaryWeight || 0}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell>Failed Checks</TableCell>
            <TableCell>{canary.status?.failedChecks || 0}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell>Iterations</TableCell>
            <TableCell>{canary.status?.iterations || 0}</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </>
  );
}

export default CanaryDetailsSection;
