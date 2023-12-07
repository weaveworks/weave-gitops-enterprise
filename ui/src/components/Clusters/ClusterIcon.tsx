import { Badge, Tooltip } from '@material-ui/core';
import { Link, formatURL } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import styled from 'styled-components';
import Azure from '../../assets/img/Azure.svg';
import Docker from '../../assets/img/docker.svg';
import EKS from '../../assets/img/EKS.svg';
import GKE from '../../assets/img/GKE.svg';
import Kubernetes from '../../assets/img/Kubernetes.svg';
import LiquidMetal from '../../assets/img/LiquidMetal.svg';
import Openshift from '../../assets/img/Openshift.svg';
import Rancher from '../../assets/img/Rancher.svg';
import VCluster from '../../assets/img/VCluster.svg';
import Vsphere from '../../assets/img/Vsphere.svg';
import { GitopsClusterEnriched } from '../../types/custom';
import { Routes } from '../../utils/nav';

const IconSpan = styled.span`
  display: flex;
  img {
    height: 32px;
    width: 32px;
  }
`;

const getClusterTypeIcon = (clusterType?: string) => {
  if (clusterType === 'DockerCluster') {
    return Docker;
  } else if (
    clusterType === 'AWSCluster' ||
    clusterType === 'AWSManagedCluster' ||
    clusterType === 'eks'
  ) {
    return EKS;
  } else if (
    clusterType === 'AzureCluster' ||
    clusterType === 'AzureManagedCluster' ||
    clusterType === 'aks'
  ) {
    return Azure;
  } else if (clusterType === 'GCPCluster') {
    return GKE;
  } else if (clusterType === 'VSphereCluster') {
    return Vsphere;
  } else if (clusterType === 'MicrovmCluster') {
    return LiquidMetal;
  } else if (clusterType === 'Rancher') {
    return Rancher;
  } else if (clusterType === 'Openshift') {
    return Openshift;
  } else if (clusterType === 'VCluster') {
    return VCluster;
  }
  return Kubernetes;
};

export const ClusterIcon: FC<{ cluster: GitopsClusterEnriched }> = ({
  cluster,
}) => {
  const clusterKind =
    cluster.annotations?.['weave.works/cluster-kind'] ||
    cluster.capiCluster?.infrastructureRef?.kind ||
    cluster.labels?.['clusters.weave.works/origin-type'];

  const isACD =
    cluster.labels?.['app.kubernetes.io/managed-by'] ===
    'cluster-reflector-controller';

  const getACDLink = () => {
    const url = formatURL(Routes.ClusterDiscoveryDetails, {
      name: cluster?.labels?.['clusters.weave.works/origin-name'],
      namespace: cluster?.labels?.['clusters.weave.works/origin-namespace'],
      clusterName: 'management',
    });
    return url;
  };

  return (
    <Tooltip title={clusterKind || 'kubernetes'} placement="bottom">
      {isACD ? (
        <Link to={getACDLink()}>
          <Badge
            anchorOrigin={{
              vertical: 'bottom',
              horizontal: 'right',
            }}
            badgeContent="ACD"
            color="primary"
          >
            <IconSpan>
              <img
                src={getClusterTypeIcon(clusterKind)}
                alt={clusterKind || 'kubernetes'}
              />
            </IconSpan>
          </Badge>
        </Link>
      ) : (
        <IconSpan>
          <img
            src={getClusterTypeIcon(clusterKind)}
            alt={clusterKind || 'kubernetes'}
          />
        </IconSpan>
      )}
    </Tooltip>
  );
};
