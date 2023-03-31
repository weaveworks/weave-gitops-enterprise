import {
    Checkbox,
    createStyles,
    makeStyles,
    withStyles
} from '@material-ui/core';
import Octicon, { Icon as ReactIcon } from '@primer/octicons-react';
import {
    DataTable,
    filterByStatusCallback,
    filterConfig, KubeStatusIndicator, statusSortHelper, theme
} from '@weaveworks/weave-gitops';
import { Condition } from '@weaveworks/weave-gitops/ui/lib/api/core/types.pb';
import React, { FC, useCallback, useState } from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { ClusterNamespacedName } from '../../cluster-services/cluster_services.pb';
import useClusters from '../../hooks/clusters';
import { GitopsClusterEnriched } from '../../types/custom';
import {
    EKSDefault,
    GKEDefault,
    KindIcon,
    Kubernetes,
    LiquidMetal,
    Openshift,
    OtherOnprem,
    Rancher,
    Vsphere
} from '../../utils/icons';
import { TableWrapper, Tooltip } from '../Shared';
import { EditButton } from '../Templates/Edit/EditButton';
import LoadingWrapper from '../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';
import { DashboardsList } from './DashboardsList';

const getClusterTypeIcon = (clusterType?: string): ReactIcon => {
  if (clusterType === 'DockerCluster') {
    return KindIcon;
  } else if (
    clusterType === 'AWSCluster' ||
    clusterType === 'AWSManagedCluster'
  ) {
    return EKSDefault;
  } else if (
    clusterType === 'AzureCluster' ||
    clusterType === 'AzureManagedCluster'
  ) {
    return Kubernetes;
  } else if (clusterType === 'GCPCluster') {
    return GKEDefault;
  } else if (clusterType === 'VSphereCluster') {
    return Vsphere;
  } else if (clusterType === 'MicrovmCluster') {
    return LiquidMetal;
  } else if (clusterType === 'Rancher') {
    return Rancher;
  } else if (clusterType === 'Openshift') {
    return Openshift;
  } else if (clusterType === 'OtherOnprem') {
    return OtherOnprem;
  }
  return Kubernetes;
};
const ClustersTableWrapper = styled(TableWrapper)`
  thead {
    th:first-of-type {
      padding: ${({ theme }) => theme.spacing.xs}
        ${({ theme }) => theme.spacing.base};
    }
  }
  td:first-of-type {
    text-overflow: clip;
    width: 25px;
    padding-left: ${({ theme }) => theme.spacing.base};
  }
  td:nth-child(7) {
    white-space: pre-wrap;
    overflow-wrap: break-word;
    word-wrap: break-word;
  }
  a {
    color: ${({ theme }) => theme.colors.primary};
  }
  width: 100%;
`;


export function computeMessage(conditions: Condition[]) {
  const readyCondition = conditions.find(
    c => c.type === 'Ready' || c.type === 'Available',
  );

  return readyCondition ? readyCondition.message : 'unknown error';
}

const useStyles = makeStyles(() =>
  createStyles({
    clusterIcon: {
      marginRight: theme.spacing.small,
      color: theme.colors.neutral30,
    },
    externalIcon: {
      marginRight: theme.spacing.small,
    },
  }),
);

export const ClusterIcon: FC<{ cluster: GitopsClusterEnriched }> = ({
  cluster,
}) => {
  const classes = useStyles();
  const clusterKind =
    cluster.labels?.['weave.works/cluster-kind'] ||
    cluster.capiCluster?.infrastructureRef?.kind;

  return (
    <Tooltip title={clusterKind || 'kubernetes'} placement="bottom">
      <span>
        <Octicon
          className={classes.clusterIcon}
          icon={getClusterTypeIcon(clusterKind)}
          size="medium"
          verticalAlign="middle"
        />
      </span>
    </Tooltip>
  );
};

const IndividualCheckbox = withStyles({
  root: {
    color: theme.colors.primary,
    '&$checked': {
      color: theme.colors.primary,
    },
    '&$disabled': {
      color: theme.colors.neutral20,
    },
  },
  checked: {},
  disabled: {},
})(Checkbox);

const ClusterRowCheckbox = ({
  name,
  namespace,
  checked,
  onChange,
}: ClusterNamespacedName & { checked: boolean; onChange: any }) => (
  <IndividualCheckbox
    checked={checked}
    onChange={useCallback(
      ev => onChange({ name, namespace }, ev),
      [name, namespace, onChange],
    )}
    name={name}
  />
);

const ClustersList = () => {
  const { clusters, isLoading } = useClusters();
  const [selectedCluster, setSelectedCluster] =
    useState<ClusterNamespacedName | null>(null);
  const initialFilterState = {
    ...filterConfig(clusters, 'status', filterByStatusCallback),
    ...filterConfig(clusters, 'namespace'),
    ...filterConfig(clusters, 'name'),
  };
  const handleIndividualClick = useCallback(
    (
      { name, namespace }: ClusterNamespacedName,
      event: React.ChangeEvent<HTMLInputElement>,
    ) => {
      if (event.target.checked === true) {
        setSelectedCluster({ name, namespace });
      } else {
        setSelectedCluster(null);
      }
    },
    [setSelectedCluster],
  );
  return (
    <LoadingWrapper loading={isLoading}>
      <ClustersTableWrapper id="clusters-list">
        <DataTable
          key={clusters.length}
          filters={initialFilterState}
          rows={clusters}
          fields={[
            {
              label: 'Select',
              value: ({ name, namespace }: GitopsClusterEnriched) => (
                <ClusterRowCheckbox
                  name={name}
                  namespace={namespace}
                  onChange={handleIndividualClick}
                  checked={Boolean(
                    selectedCluster?.name === name &&
                      selectedCluster?.namespace === namespace,
                  )}
                />
              ),
              maxWidth: 25,
            },
            {
              label: 'Name',
              value: (c: GitopsClusterEnriched) =>
                c.controlPlane === true ? (
                  <span data-cluster-name={c.name}>{c.name}</span>
                ) : (
                  <Link
                    to={`/cluster?clusterName=${c.name}`}
                    color={theme.colors.primary}
                    data-cluster-name={c.name}
                  >
                    {c.name}
                  </Link>
                ),
              sortValue: ({ name }) => name,
              textSearchable: true,
              maxWidth: 275,
            },
            {
              label: 'Dashboards',
              value: (c: GitopsClusterEnriched) => (
                <DashboardsList cluster={c} />
              ),
            },
            {
              label: 'Type',
              value: (c: GitopsClusterEnriched) => (
                <ClusterIcon cluster={c}></ClusterIcon>
              ),
            },
            {
              label: 'Namespace',
              value: 'namespace',
            },
            {
              label: 'Status',
              value: (c: GitopsClusterEnriched) =>
                c.conditions && c.conditions.length > 0 ? (
                  <KubeStatusIndicator short conditions={c.conditions} />
                ) : null,
              sortValue: statusSortHelper,
            },
            {
              label: 'Message',
              value: (c: GitopsClusterEnriched) =>
                (c.conditions && c.conditions[0]?.message) || null,
              sortValue: ({ conditions }) => computeMessage(conditions),
              maxWidth: 600,
            },
            {
              label: '',
              value: (c: GitopsClusterEnriched) => <EditButton resource={c} />,
            },
          ]}
        />
      </ClustersTableWrapper>
    </LoadingWrapper>
  );
};

export default ClustersList;
