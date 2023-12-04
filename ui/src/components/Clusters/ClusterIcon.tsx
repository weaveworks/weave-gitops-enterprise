import { FC } from 'react';
import { GitopsClusterEnriched } from '../../types/custom';
import { Badge, Tooltip } from '@material-ui/core';
import { Link } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import Docker from '../../assets/img/docker.svg';
import EKS from '../../assets/img/EKS.svg';
import GKE from '../../assets/img/GKE.svg';
import Kubernetes from '../../assets/img/Kubernetes.svg';
import LiquidMetal from '../../assets/img/LiquidMetal.svg';
import Openshift from '../../assets/img/Openshift.svg';
import Rancher from '../../assets/img/Rancher.svg';
import Vsphere from '../../assets/img/Vsphere.svg';



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
      clusterType === 'AWSManagedCluster'
    ) {
      return EKS;
    } else if (
      clusterType === 'AzureCluster' ||
      clusterType === 'AzureManagedCluster'
    ) {
      return Kubernetes;
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
    }
    return Kubernetes;
  };
  


export const ClusterIcon: FC<{ cluster: GitopsClusterEnriched }> = ({
    cluster,
  }) => {
    const clusterKind =
      cluster.labels?.['weave.works/cluster-kind'] ||
      cluster.capiCluster?.infrastructureRef?.kind;
  
    const isACD =
      cluster.labels?.['app.kubernetes.io/managed-by'] ===
      'cluster-reflector-controller'
        ? true
        : false;
    return (
      <Tooltip title={clusterKind || 'kubernetes'} placement="bottom">
        {isACD ? (
          <Badge
            anchorOrigin={{
              vertical: 'bottom',
              horizontal: 'right',
            }}
            badgeContent="ACD"
            color="primary"
          >
            <Link
              to={`/cluster-discovery/object/details?clusterName=management&name=${cluster?.labels?.['clusters.weave.works/origin-name']}&namespace=${cluster?.labels?.['clusters.weave.works/origin-namespace']}`}
            >
              <IconSpan>
                <img
                  src={getClusterTypeIcon(clusterKind)}
                  alt={clusterKind || 'kubernetes'}
                />
              </IconSpan>
            </Link>
          </Badge>
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