import { Box, Button, Typography } from '@material-ui/core';
import Chip from '@material-ui/core/Chip';
import { InfoList, KubeStatusIndicator, theme } from '@weaveworks/weave-gitops';
import { InfoField } from '@weaveworks/weave-gitops/ui/components/InfoList';
import { useState } from 'react';
import styled from 'styled-components';
import { GitopsClusterEnriched } from '../../types/custom';
import { DashboardsList } from './DashboardsList';
import { ClusterStatus } from './ClusterStatus';

export const sectionTitle = (title: string) => (
  <Typography
    style={{ fontWeight: 'bold', marginTop: theme.spacing.medium }}
    variant="h6"
    gutterBottom
    component="div"
  >
    {title}
  </Typography>
);

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
  const capiClusterAnnotations = currentCluster?.capiCluster?.annotations || {};
  const capiClusterLabels = currentCluster?.capiCluster?.labels || {};
  const infrastructureRef = currentCluster?.capiCluster?.infrastructureRef;
  const [disabled, setDisabled] = useState<boolean>(false);
  const dashboardAnnotations = getDashboardAnnotations(
    currentCluster as GitopsClusterEnriched,
  );

  console.log(currentCluster);

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
    labels: GitopsClusterEnriched['labels'] | null,
    annotations: GitopsClusterEnriched['annotations'] | null,
  ) => {
    const getObjects = () => {
      if (labels) return Object.entries(labels);
      if (annotations) return Object.entries(annotations);
      return [];
    };
    return (
      <Box>
        <Typography variant="body1" gutterBottom component="div">
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
        <Box>
          <InfoList items={info as [string, any][]} />
        </Box>
        {Object.keys(dashboardAnnotations).length > 0 && (
          <Box margin={2}>
            <Typography variant="h6" gutterBottom component="div">
              Dashboards
            </Typography>
            <DashboardsList cluster={currentCluster as GitopsClusterEnriched} />
          </Box>
        )}

        {/* GitOpsCluster */}
        {sectionTitle('GitOps Cluster')}
        {Object.keys(labels).length > 0 && renderer(labels, null)}
        {Object.keys(annotations).length > 0 && renderer(null, annotations)}
        <Box>
          <ClusterStatus
            clusterName={currentCluster.name}
            conditions={currentCluster?.conditions}
          />
        </Box>

        {/* CapiCluster */}
        {sectionTitle('CAPI Cluster')}
        <Box>
          <InfoList items={[['Name', currentCluster?.capiCluster?.name]]} />
        </Box>
        {Object.keys(capiClusterLabels).length > 0 &&
          renderer(capiClusterLabels, null)}
        {Object.keys(capiClusterAnnotations).length > 0 &&
          renderer(null, capiClusterAnnotations)}
        <Box>
          <ClusterStatus
            clusterName={currentCluster.name}
            status={currentCluster?.capiCluster?.status}
          />
        </Box>
        {infrastructureRef && (
          <Box>
            {sectionTitle('Infrastructure')}
            <InfoList items={infrastructureRefInfo} />
          </Box>
        )}
      </ClusterDashbordWrapper>
    )
  );
};

export default ClusterDashboard;
