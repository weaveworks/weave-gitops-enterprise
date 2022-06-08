import React, { FC, useEffect, useState } from 'react';
import { ThemeProvider } from '@material-ui/core/styles';
import useClusters from '../../contexts/Clusters';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { useParams, useRouteMatch } from 'react-router-dom';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { localEEMuiTheme } from '../../muiTheme';
import { CAPIClusterStatus } from './CAPIClusterStatus';
import { GitopsClusterEnriched } from '../../types/custom';
import {
  EventsTable,
  Flex,
  FluxObjectKind,
  InfoList,
  Metadata,
  RouterTab,
  SubRouterTabs,
} from '@weaveworks/weave-gitops';
import ListItem from '@material-ui/core/ListItem';
import { Box, Button, Typography } from '@material-ui/core';
import { DashboardsList } from './DashboardsList';

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

const ClusterDashboard = ({
  className,
  name,
  namespace,
  clusterName,
}: Props) => {
  const { getCluster, getKubeconfig, count } = useClusters();
  // const { clusterName } = useParams<{ clusterName: string }>();

  const [currentCluster, setCurrentCluster] =
    useState<GitopsClusterEnriched | null>(null);
  const { path } = useRouteMatch();

  const handleClick = () =>
    getKubeconfig(clusterName, `${clusterName}.kubeconfig`);

  const info = [
    [
      'kubeconfig',
      <Button
        // className={classes.downloadBtn}
        onClick={handleClick}
      >
        Download the kubeconfig here
      </Button>,
    ],
    ['Namespace', currentCluster?.namespace],
  ];

  console.log(clusterName);

  console.log('path', path);

  useEffect(() => {
    // if (clusterName) {
    setCurrentCluster(getCluster(clusterName));
    // }
  }, [clusterName, getCluster]);

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
          <Flex wide tall column>
            <SubRouterTabs rootPath={`${path}/details`}>
              <RouterTab name="Details" path={`${path}/details`}>
                <>
                  <InfoList items={info as [string, any][]} />
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
                        <ListItem key={key}>
                          {key}: {value}
                        </ListItem>
                      ),
                    )}
                  </Box>
                  <CAPIClusterStatus
                    clusterName={clusterName}
                    status={currentCluster?.capiCluster?.status}
                  />
                </>
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
          </Flex>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default ClusterDashboard;
