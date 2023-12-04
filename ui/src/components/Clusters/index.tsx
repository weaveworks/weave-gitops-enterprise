import { Badge, Checkbox } from '@material-ui/core';
import {
  Button,
  DataTable,
  Flex,
  GitRepository,
  Icon,
  IconType,
  Kind,
  KubeStatusIndicator,
  Link,
  filterByStatusCallback,
  filterConfig,
  statusSortHelper,
  useListSources,
} from '@weaveworks/weave-gitops';
import { Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import _ from 'lodash';
import React, { FC, useCallback, useEffect, useMemo, useState } from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { GitProvider } from '../../api/gitauth/gitauth.pb';
import { EnabledComponent } from '../../api/query/query.pb';
import Docker from '../../assets/img/docker.svg';
import EKS from '../../assets/img/EKS.svg';
import GKE from '../../assets/img/GKE.svg';
import Kubernetes from '../../assets/img/Kubernetes.svg';
import LiquidMetal from '../../assets/img/LiquidMetal.svg';
import Openshift from '../../assets/img/Openshift.svg';
import Rancher from '../../assets/img/Rancher.svg';
import Vsphere from '../../assets/img/Vsphere.svg';
import { ClusterNamespacedName } from '../../cluster-services/cluster_services.pb';
import CallbackStateContextProvider from '../../contexts/GitAuth/CallbackStateContext';
import { useListConfigContext } from '../../contexts/ListConfig';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';
import useClusters from '../../hooks/clusters';
import { useIsEnabledForComponent } from '../../hooks/query';
import AppRoutes from '../../routes';
import { GitopsClusterEnriched, PRDefaults } from '../../types/custom';
import { useCallbackState } from '../../utils/callback-state';
import { computeMessage } from '../../utils/conditions';
import { toFilterQueryString } from '../../utils/FilterQueryString';
import { Routes } from '../../utils/nav';
import { QueryState } from '../Explorer/hooks';
import { linkToExplorer } from '../Explorer/utils';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { Tooltip } from '../Shared';
import { EditButton } from '../Templates/Edit/EditButton';
import {
  getCreateRequestAnnotation,
  useGetInitialGitRepo,
} from '../Templates/Form/utils';
import LoadingWrapper from '../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';
import { ConnectClusterDialog } from './ConnectInfoBox';
import { DashboardsList } from './DashboardsList';
import { DeleteClusterDialog } from './Delete';
import OpenedPullRequest from './OpenedPullRequest';

const IconSpan = styled.span`
  display: flex;
  img {
    height: 32px;
    width: 32px;
  }
`;

export const ClusterIcon: FC<{ cluster: GitopsClusterEnriched }> = ({
  cluster,
}) => {
  const clusterKind =
    cluster.labels?.['weave.works/cluster-kind'] ||
    cluster.capiCluster?.infrastructureRef?.kind;

  const isACD = cluster;
  return (
    <Tooltip title={clusterKind || 'kubernetes'} placement="bottom">
      <Badge badgeContent="ACD" color="primary" invisible={false}>
        <IconSpan>
          <img
            src={getClusterTypeIcon(clusterKind)}
            alt={clusterKind || 'kubernetes'}
          />
        </IconSpan>
      </Badge>
    </Tooltip>
  );
};

const ClusterRowCheckbox = ({
  name,
  namespace,
  checked,
  onChange,
}: ClusterNamespacedName & { checked: boolean; onChange: any }) => (
  <Checkbox
    checked={checked}
    color="primary"
    onChange={useCallback(
      ev => onChange({ name, namespace }, ev),
      [name, namespace, onChange],
    )}
    name={name}
  />
);

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

interface FormData {
  repo: GitRepository | null;
  branchName: string;
  pullRequestTitle: string;
  commitMessage: string;
  pullRequestDescription: string;
}

export const getGitRepos = (sources: Source[] | undefined) =>
  _.orderBy(
    _.uniqBy(
      _.filter(
        sources,
        (item): item is GitRepository => item.type === Kind.GitRepository,
      ),
      repo => repo?.obj?.spec?.url,
    ),
    ['url'],
    ['asc'],
  );

const MCCP: FC<{
  location: { state: { notification: NotificationData[] } };
}> = ({ location }) => {
  const { clusters, isLoading } = useClusters();
  const { setNotifications } = useNotifications();
  const [selectedCluster, setSelectedCluster] =
    useState<ClusterNamespacedName | null>(null);
  const [openConnectInfo, setOpenConnectInfo] = useState<boolean>(false);
  const [openDeletePR, setOpenDeletePR] = useState<boolean>(false);
  const handleClose = useCallback(() => {
    setOpenDeletePR(false);
    setSelectedCluster(null);
  }, [setOpenDeletePR, setSelectedCluster]);
  const { data: sources } = useListSources();
  const isExplorerEnabled = useIsEnabledForComponent(
    EnabledComponent.templates,
  );

  const gitRepos = useMemo(
    () => getGitRepos(sources?.result),
    [sources?.result],
  );
  const listConfigContext = useListConfigContext();
  const provider = listConfigContext?.provider;

  const capiClusters = useMemo(
    () => clusters.filter(cls => cls.capiCluster),
    [clusters],
  );
  let selectedCapiCluster = useMemo(
    () =>
      capiClusters.find(
        c =>
          c.name === selectedCluster?.name &&
          c.namespace === selectedCluster?.namespace,
      ) || null,
    [capiClusters, selectedCluster],
  );
  const [random, setRandom] = useState<string>(
    Math.random().toString(36).substring(7),
  );

  useEffect(() => {
    if (openDeletePR === true) {
      setRandom(Math.random().toString(36).substring(7));
    }
  }, [openDeletePR]);

  const PRdefaults: PRDefaults = {
    branchName: `delete-clusters-branch-${random}`,
    pullRequestTitle: 'Deletes capi cluster(s)',
    commitMessage: 'Deletes capi cluster(s)',
  };

  const authRedirectPage = `/clusters`;

  let initialFormData = {
    ...PRdefaults,
    repo: null,
    pullRequestDescription: '',
  };

  const callbackState = useCallbackState();

  if (callbackState) {
    initialFormData = {
      ...initialFormData,
      ...callbackState.state.formData,
    };
    selectedCapiCluster = {
      ...selectedCapiCluster,
      ...callbackState.state.selectedCapiCluster,
    };
  }

  const [formData, setFormData] = useState<FormData>(initialFormData);
  const initialUrl =
    selectedCapiCluster &&
    getCreateRequestAnnotation(selectedCapiCluster)?.repository_url;
  const initialGitRepo = useGetInitialGitRepo(initialUrl, gitRepos);

  const history = useHistory();

  const handleAddCluster = useCallback(() => {
    if (isExplorerEnabled) {
      const url = linkToExplorer(Routes.Templates, {
        filters: [`labels.weave.works/template-type:cluster`],
      } as QueryState);
      // Explorer uses a different query param for filters to avoid conflicts with DataTable
      history.push(url);
    } else {
      const filtersValues = toFilterQueryString([
        { key: 'templateType', value: 'cluster' },
        { key: 'templateType', value: '' },
      ]);
      history.push(`/templates?filters=${filtersValues}`);
    }
  }, [history, isExplorerEnabled]);

  const initialFilterState = {
    ...filterConfig(clusters, 'status', filterByStatusCallback),
    ...filterConfig(clusters, 'namespace'),
    ...filterConfig(clusters, 'name'),
  };

  useEffect(() => {
    if (!callbackState) {
      const prTitle = `Delete cluster: ${selectedCluster?.namespace}/${selectedCluster?.name}`;
      setFormData((prevState: FormData) => ({
        ...prevState,
        gitRepos,
        commitMessage: prTitle,
        pullRequestTitle: prTitle,
        pullRequestDescription: prTitle,
      }));
    }

    if (!callbackState && !selectedCluster) {
      setOpenDeletePR(false);
    }

    if (callbackState?.state?.selectedCapiCluster) {
      setOpenDeletePR(true);
    }
  }, [
    callbackState,
    selectedCapiCluster,
    capiClusters,
    selectedCluster,
    gitRepos,
  ]);

  useEffect(
    () =>
      setNotifications([
        {
          message: {
            text: location?.state?.notification?.[0]?.message.text,
          },
          severity: location?.state?.notification?.[0]?.severity,
        } as NotificationData,
      ]),
    [location?.state?.notification, setNotifications],
  );

  useEffect(() => {
    if (!formData.repo) {
      setFormData((prevState: any) => ({
        ...prevState,
        repo: initialGitRepo,
      }));
    }
  }, [initialGitRepo, formData.repo]);

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
    <Page path={[{ label: 'Clusters' }]}>
      <CallbackStateContextProvider
        callbackState={{
          page: authRedirectPage as PageRoute,
          state: { formData, selectedCapiCluster },
        }}
      >
        <NotificationsWrapper>
          <Flex column gap="24">
            <Flex gap="12">
              <Button
                id="create-cluster"
                startIcon={<Icon type={IconType.AddIcon} size="base" />}
                onClick={handleAddCluster}
              >
                CREATE A CLUSTER
              </Button>
              <Button
                id="connect-cluster"
                startIcon={<Icon type={IconType.ArrowUpwardIcon} size="base" />}
                onClick={() => setOpenConnectInfo(true)}
              >
                CONNECT A CLUSTER
              </Button>
              <Tooltip
                title={
                  provider === GitProvider.BitBucketServer
                    ? 'Operation is not supported'
                    : 'No CAPI cluster selected'
                }
                placement="top"
                disabled={
                  Boolean(selectedCapiCluster) &&
                  provider !== GitProvider.BitBucketServer
                }
              >
                <div>
                  <Button
                    id="delete-cluster"
                    startIcon={<Icon type={IconType.DeleteIcon} size="base" />}
                    onClick={() => {
                      setNotifications([]);
                      setOpenDeletePR(true);
                    }}
                    disabled={
                      !selectedCapiCluster ||
                      provider === GitProvider.BitBucketServer
                    }
                  >
                    CREATE A PR TO DELETE CLUSTERS
                  </Button>
                </div>
              </Tooltip>
              {openDeletePR && (
                <DeleteClusterDialog
                  formData={formData}
                  setFormData={setFormData}
                  selectedCapiCluster={
                    selectedCapiCluster || ({} as ClusterNamespacedName)
                  }
                  onClose={handleClose}
                  prDefaults={PRdefaults}
                />
              )}
              {openConnectInfo && (
                <ConnectClusterDialog
                  onFinish={() => setOpenConnectInfo(false)}
                />
              )}
              <OpenedPullRequest />
            </Flex>
            <LoadingWrapper loading={isLoading}>
              <DataTable
                className="clusters-list"
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
                          to={`/cluster?clusterName=${c.name}&namespace=${c.namespace}`}
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
                    value: (c: GitopsClusterEnriched) => (
                      <EditButton resource={c} />
                    ),
                  },
                ]}
              />
            </LoadingWrapper>
          </Flex>
        </NotificationsWrapper>
      </CallbackStateContextProvider>
    </Page>
  );
};

export default MCCP;
