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
import {
  RouterTab,
  SubRouterTabs,
  EventsTable,
  FluxObjectKind,
} from '@weaveworks/weave-gitops';

import {
  getDeploymentStrategyIcon,
  getProgressValue,
} from '../ListCanaries/Table';
import {
  Canary,
  Automation,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { darcula } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { useRouteMatch } from 'react-router-dom';

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
  automation: Automation;
}) {
  const classes = useCanaryStyle();
  const { path } = useRouteMatch();

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

            <div
              className={`${classes.sectionHeaderWrapper} ${classes.cardTitle}`}
            >
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
          </CanaryDetailsWrapper>
        </RouterTab>
        <RouterTab name="yaml" path={`${path}/yaml`}>
          <CanaryDetailsWrapper>
            <SyntaxHighlighter
              language="yaml"
              style={darcula}
              wrapLongLines="pre-wrap"
              showLineNumbers={true}
              codeTagProps={{
                className: classes.code,
              }}
              customStyle={{
                height: '450px',
              }}
            >
              {canary.yaml}
            </SyntaxHighlighter>
          </CanaryDetailsWrapper>
        </RouterTab>
        <RouterTab name="Events" path={`${path}/events`}>
          <CanaryDetailsWrapper>
            <EventsTable
              namespace={canary?.namespace}
              involvedObject={{
                kind: FluxObjectKind.KindCluster,
                name: canary.clusterName,
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