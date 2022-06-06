import React, { FC, useEffect, useState } from 'react';
import { ThemeProvider } from '@material-ui/core/styles';
import useClusters from '../../contexts/Clusters';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { useParams } from 'react-router-dom';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { localEEMuiTheme } from '../../muiTheme';
import { CAPIClusterStatus } from './CAPIClusterStatus';
import { GitopsClusterEnriched } from '../../types/custom';
import { Box, ListItem, Typography } from '@material-ui/core';

const ClusterDashboard: FC = () => {
  const { getCluster, count } = useClusters();
  const { clusterName } = useParams<{ clusterName: string }>();
  const [currentCluster, setCurrentCluster] =
    useState<GitopsClusterEnriched | null>(null);

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
          <CAPIClusterStatus
            clusterName={clusterName}
            status={currentCluster?.capiCluster?.status}
          />
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
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default ClusterDashboard;
