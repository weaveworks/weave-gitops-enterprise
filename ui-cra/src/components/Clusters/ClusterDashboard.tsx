import React, { useEffect, useState } from 'react';
import styled from 'styled-components';
import { ThemeProvider } from '@material-ui/core/styles';
import useClusters from '../../contexts/Clusters';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { useRouteMatch } from 'react-router-dom';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { localEEMuiTheme } from '../../muiTheme';
import { CAPIClusterStatus } from './CAPIClusterStatus';
import { GitopsClusterEnriched } from '../../types/custom';
import {
  EventsTable,
  FluxObjectKind,
  InfoList,
  RouterTab,
  SubRouterTabs,
} from '@weaveworks/weave-gitops';
import { Box, Button, Typography } from '@material-ui/core';
import { DashboardsList } from './DashboardsList';
import Chip from '@material-ui/core/Chip';
import Divider from '@material-ui/core/Divider';

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

const ClusterDashbordWrapper = styled.div`
  .kubeconfig-download {
    padding: 0;
  }
`;

const ClusterDashboard = ({ clusterName }: Props) => {
  const { getCluster, getKubeconfig, count } = useClusters();
  const [currentCluster, setCurrentCluster] =
    useState<GitopsClusterEnriched | null>(null);
  const { path } = useRouteMatch();

  const handleClick = () =>
    getKubeconfig(clusterName, `${clusterName}.kubeconfig`);

  const info = [
    [
      'kubeconfig',
      <Button className="kubeconfig-download" onClick={handleClick}>
        Download the kubeconfig here
      </Button>,
    ],
    ['Namespace', currentCluster?.namespace],
  ];

  useEffect(
    () => setCurrentCluster(getCluster(clusterName)),
    [clusterName, getCluster],
  );

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Cluster Page">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Clusters', url: 'clusters', count },
            { label: clusterName },
          ]}
        />
        <ContentWrapper>
          <SubRouterTabs rootPath={`${path}/details`}>
            <RouterTab name="Details" path={`${path}/details`}>
              <ClusterDashbordWrapper>
                <InfoList items={info as [string, any][]} />
                <Divider variant="middle" />
                <Box margin={2}>
                  <Typography variant="h6" gutterBottom component="div">
                    DASHBOARDS
                  </Typography>
                  <DashboardsList
                    cluster={currentCluster as GitopsClusterEnriched}
                  />
                </Box>
                <Box margin={2}>
                  <Typography variant="h6" gutterBottom component="div">
                    LABELS
                  </Typography>
                  {Object.entries(currentCluster?.labels || {}).map(
                    ([key, value]) => (
                      <Chip key={key} label={`${key}: ${value}`} />
                    ),
                  )}
                </Box>
                <CAPIClusterStatus
                  clusterName={clusterName}
                  status={currentCluster?.capiCluster?.status}
                />
              </ClusterDashbordWrapper>
            </RouterTab>
            <RouterTab name="Events" path={`${path}/events`}>
              <EventsTable
                namespace={currentCluster?.namespace}
                involvedObject={{
                  kind: 'KindCluster' as FluxObjectKind,
                  name: currentCluster?.name,
                  namespace: currentCluster?.namespace,
                }}
              />
            </RouterTab>
          </SubRouterTabs>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default ClusterDashboard;
