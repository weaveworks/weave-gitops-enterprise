import { Box, Button, Typography } from '@material-ui/core';
import Chip from '@material-ui/core/Chip';
import Divider from '@material-ui/core/Divider';
import { InfoList, KubeStatusIndicator, theme } from '@weaveworks/weave-gitops';
import { InfoField } from '@weaveworks/weave-gitops/ui/components/InfoList';
import { useState } from 'react';
import styled from 'styled-components';
import { GitopsClusterEnriched } from '../../types/custom';

import { CAPIClusterStatus } from './CAPIClusterStatus';
import { DashboardsList } from './DashboardsList';

const ClusterDashbordWrapper = styled.div`
  .kubeconfig-download {
    padding: 0;
    font-weight: bold;
    color: ${theme.colors.primary};
  }
`;

const ClusterDashboard = ({
  currentCluster,
  getDashboardAnnotations,
  getKubeconfig,
}: {
  currentCluster: GitopsClusterEnriched;
  getDashboardAnnotations: (cluster: GitopsClusterEnriched) => {
    [key: string]: string;
  };
  getKubeconfig: (
    clusterName: string,
    clusterNamespace: string,
    filename: string,
  ) => Promise<void>;
}) => {
  const labels = currentCluster?.labels || {};
  const infrastructureRef = currentCluster?.capiCluster?.infrastructureRef;

  const [disabled, setDisabled] = useState<boolean>(false);
  const dashboardAnnotations = getDashboardAnnotations(
    currentCluster as GitopsClusterEnriched,
  );

  const handleClick = () => {
    setDisabled(true);
    getKubeconfig(
      currentCluster.name || '',
      currentCluster?.namespace || '',
      `${currentCluster?.name}.kubeconfig`,
    ).finally(() => {
      setDisabled(false);
    });
  };

  const info = [
    [
      'kubeconfig',
      <Button
        className="kubeconfig-download"
        onClick={handleClick}
        disabled={disabled}
      >
        Kubeconfig
      </Button>,
    ],
    ['Namespace', currentCluster?.namespace],
  ];

  const infrastructureRefInfo: InfoField[] = infrastructureRef
    ? [
        ['Kind', infrastructureRef.kind],
        ['APIVersion', infrastructureRef.apiVersion],
      ]
    : [];
  return (
    <>
      {currentCluster && (
        <ClusterDashbordWrapper>
          {currentCluster?.conditions &&
          currentCluster?.conditions[0]?.message ? (
            <div style={{ paddingBottom: theme.spacing.small }}>
              <KubeStatusIndicator conditions={currentCluster.conditions} />
            </div>
          ) : null}

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
              clusterName={currentCluster.name}
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
      )}
    </>
  );
};

export default ClusterDashboard;
