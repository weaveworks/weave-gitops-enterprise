import { Box, Button, Typography } from '@material-ui/core';
import Chip from '@material-ui/core/Chip';
import Divider from '@material-ui/core/Divider';
import { InfoList, KubeStatusIndicator, theme } from '@weaveworks/weave-gitops';
import { InfoField } from '@weaveworks/weave-gitops/ui/components/InfoList';
import { useState } from 'react';
import styled from 'styled-components';
import { GitopsClusterEnriched } from '../../types/custom';
import { DashboardsList } from './DashboardsList';
import { ClusterStatus } from './ClusterStatus';

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
  const annotations = currentCluster?.annotations || {};
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

  const renderer = (
    labels?: GitopsClusterEnriched['labels'],
    annotations?: GitopsClusterEnriched['annotations'],
  ) => {
    const getObjects = () => {
      if (labels) return Object.entries(labels);
      if (annotations) return Object.entries(annotations);
      return [];
    };
    return (
      <Box margin={2}>
        <Typography variant="h6" gutterBottom component="div">
          {labels && 'Labels'}
          {annotations && 'Annotations'}
        </Typography>
        {getObjects().map(([key, value]) => (
          <Chip
            title={value}
            style={{
              maxWidth: '650px',
              marginRight: theme.spacing.small,
              marginBottom: theme.spacing.small,
            }}
            key={key}
            label={`${key}: ${value}`}
          />
        ))}
      </Box>
    );
  };

  return (
    currentCluster && (
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
            {renderer(labels)} <Divider variant="middle" />
          </>
        ) : null}
        {Object.keys(annotations).length > 0 ? (
          <>
            {renderer(annotations)} <Divider variant="middle" />
          </>
        ) : null}
        <Box margin={2}>
          <ClusterStatus
            clusterName={currentCluster.name}
            conditions={currentCluster?.conditions}
          />
        </Box>
        <Divider variant="middle" />
        <Box margin={2}>
          <ClusterStatus
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
    )
  );
};

export default ClusterDashboard;
