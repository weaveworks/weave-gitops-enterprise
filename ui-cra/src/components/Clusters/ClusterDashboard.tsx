import React, { useEffect, useState } from 'react';
import styled from 'styled-components';
import { ThemeProvider } from '@material-ui/core/styles';
import useClusters from '../../contexts/Clusters';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { useHistory, useRouteMatch } from 'react-router-dom';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { localEEMuiTheme } from '../../muiTheme';
import { CAPIClusterStatus } from './CAPIClusterStatus';
import { GitopsClusterEnriched } from '../../types/custom';
import {
  theme,
  InfoList,
  RouterTab,
  SubRouterTabs,
  Icon,
  IconType,
  Button as WeaveButton,
  KubeStatusIndicator,
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
  const { getCluster, getDashboardAnnotations, getKubeconfig, count } =
    useClusters();
  const [currentCluster, setCurrentCluster] =
    useState<GitopsClusterEnriched | null>(null);
  const { path } = useRouteMatch();
  const labels = currentCluster?.labels || {};
  const infrastructureRef = currentCluster?.capiCluster?.infrastructureRef;
  const dashboardAnnotations = getDashboardAnnotations(
    currentCluster as GitopsClusterEnriched,
  );
  const history = useHistory();

  const handleClick = () =>
    getKubeconfig(
      clusterName,
      currentCluster?.namespace || '',
      `${clusterName}.kubeconfig`,
    );

  const info = [
    [
      'kubeconfig',
      <Button className="kubeconfig-download" onClick={handleClick}>
        Download the kubeconfig here
      </Button>,
    ],
    ['Namespace', currentCluster?.namespace],
  ];

  const infrastructureRefInfo: InfoField[] = infrastructureRef ? [['Kind', infrastructureRef.kind]] :[];

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
            { label: 'Clusters', url: '/clusters', count },
            { label: clusterName },
          ]}
        />
        <ContentWrapper>
          <SubRouterTabs rootPath={`${path}/details`}>
            <RouterTab name="Details" path={`${path}/details`}>
              <ClusterDashbordWrapper>
                {currentCluster?.conditions &&
                currentCluster?.conditions[0]?.message ? (
                  <div style={{ paddingBottom: theme.spacing.small }}>
                    <KubeStatusIndicator
                      conditions={currentCluster.conditions}
                    />
                  </div>
                ) : null}

                <WeaveButton
                  id="create-cluster"
                  startIcon={<Icon type={IconType.ExternalTab} size="base" />}
                  onClick={() => {
                    const filtersValues = encodeURIComponent(
                      `clusterName:${currentCluster?.namespace}/${currentCluster?.name}`,
                    );
                    history.push(`/applications?filters=${filtersValues}`);
                  }}
                >
                  GO TO APPLICATIONS
                </WeaveButton>
                <Box margin={2}>
                  <InfoList items={info as [string, any][]} />
                </Box>
                <Divider variant="middle" />
                {Object.keys(dashboardAnnotations).length > 0 ? (
                  <>
                    <Box margin={2}>
                      <Typography variant="h6" gutterBottom component="div">
                        Dashboards
                      </Typography>
                      <DashboardsList
                        cluster={currentCluster as GitopsClusterEnriched}
                      />
                    </Box>
                    <Divider variant="middle" />
                  </>
                ) : null}
                {Object.keys(labels).length > 0 ? (
                  <>
                    <Box margin={2}>
                      <Typography variant="h6" gutterBottom component="div">
                        Labels
                      </Typography>
                      {Object.entries(labels).map(([key, value]) => (
                        <Chip
                          style={{ marginRight: theme.spacing.small }}
                          key={key}
                          label={`${key}: ${value}`}
                        />
                      ))}
                    </Box>
                    <Divider variant="middle" />
                  </>
                ) : null}
                <Box margin={2}>
                  <CAPIClusterStatus
                    clusterName={clusterName}
                    status={currentCluster?.capiCluster?.status}
                  />
                </Box>
                {infrastructureRef ? (
                  <>
                  <Divider variant="middle" />
                    <Box margin={2}>
                      <Typography variant="h6" gutterBottom component="div">
                        Infrastructure
                      </Typography>
                      <InfoList items={infrastructureRefInfo} />
                    </Box>
                  </>
                ) : null}
              </ClusterDashbordWrapper>
            </RouterTab>
          </SubRouterTabs>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default ClusterDashboard;
