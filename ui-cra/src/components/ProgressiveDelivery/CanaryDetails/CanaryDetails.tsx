import React, { useCallback, useState } from 'react';
// import { useParams } from 'react-router-dom';
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
// import { useParams } from 'react-router-dom';
import styled from 'styled-components';
import { ProgressiveDeliveryService } from '../../../cluster-services/prog.pb';

const TitleWrapper = styled.h2`
  margin: 0px;
`;

function CanaryDetails() {
  // const { id } = useParams<{ id: string }>();
  const [name, setName] = useState('');
  const fetchPoliciesAPI = useCallback(
    () =>
      ProgressiveDeliveryService.GetCanary({}).then((res: any) => {
        res.canary && setName(res.canary?.name || '');
        return res;
      }),
    [],
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
            {({ value: { canary } }: { value: { canary: any } }) => (
              <>
                <TitleWrapper>{name}</TitleWrapper>
                <div className={classes.statusWrapper}>
                  <CanaryStatus
                    status={canary.status.phase}
                    canaryWeight={canary.status.canaryWeight || 0}
                  />
                  <p className={classes.statusMessage}>
                    {canary.status.conditions[0].message}
                  </p>
                </div>
                <CanaryRowHeader rowkey="Cluster" value={canary.clusterName} />
                <CanaryRowHeader rowkey="Namespace" value={canary.namespace} />
                <CanaryRowHeader
                  rowkey="Target"
                  value={`${canary.targetDeployment.kind}/${canary.targetDeployment.name}`}
                />
                <CanaryRowHeader rowkey="Service" value="-" />
                <CanaryRowHeader rowkey="Provider" value={canary.provider} />
                <CanaryRowHeader
                  rowkey="Last Transition Time"
                  value={canary.status.lastTransitionTime}
                />
                {/* <CanaryRowHeader rowkey="Last Transition Time" value={ moment(canary.status.lastUpdateTime).fromNow()} /> */}
                <CanaryRowHeader
                  rowkey="Last Updated Time"
                  value={canary.status.conditions[0].lastUpdateTime}
                />

                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell align="left">
                        <span>Status</span>
                      </TableCell>
                      <TableCell align="left">
                        <span>Value</span>
                      </TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    <TableRow>
                      <TableCell>canaryWeight</TableCell>
                      <TableCell>{canary.status?.canaryWeight || 0}</TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell>failedChecks</TableCell>
                      <TableCell>{canary.status?.failedChecks || 0}</TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell>iterationa</TableCell>
                      <TableCell>{canary.status?.iterationa || 0}</TableCell>
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
