import { useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';

import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import LoadingError from '../../LoadingError';
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
import { ProgressiveDeliveryService } from '../../../cluster-services/prog.pb';
import { Canary } from '../../../cluster-services/types.pb';
import { Automation } from '@weaveworks/weave-gitops';
import { getDeploymentStrategyIcon } from '../ListCanaries/Table';

const TitleWrapper = styled.h2`
  margin: 0px;
`;
const SectionHeaderWrapper = styled.div`
  margin: 0px;
`;

interface ICanaryParams {
  name: string;
  namespace: string;
  clusterName: string;
}
interface IGetCanaryReponse {
  canary: Canary;
  automation: Automation;
}
function CanaryDetails() {
  const { name, namespace, clusterName } = useParams<ICanaryParams>();
  const fetchPoliciesAPI = useCallback(
    () =>
      ProgressiveDeliveryService.GetCanary({
        name,
        namespace,
        clusterName,
      }),
    [clusterName, name, namespace],
  );
  const classes = useCanaryStyle();

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Policies">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Applications', url: '/applications' },
            { label: 'Delivery', url: '/applications/delivery' },
            { label: name, url: 'canary-details' },
          ]}
        />
        <ContentWrapper>
          <LoadingError fetchFn={fetchPoliciesAPI}>
            {({
              value: { canary, automation },
            }: {
              value: IGetCanaryReponse;
            }) => (
              <>
                <TitleWrapper>{name}</TitleWrapper>
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
              </>
            )}
          </LoadingError>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
}

export default CanaryDetails;
