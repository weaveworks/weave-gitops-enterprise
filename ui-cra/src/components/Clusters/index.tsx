import {
  Checkbox,
  createStyles,
  makeStyles,
  withStyles,
} from '@material-ui/core';
import EditIcon from '@material-ui/icons/Edit';
import Octicon, { Icon as ReactIcon } from '@primer/octicons-react';
import {
  Button,
  CallbackStateContextProvider,
  filterByStatusCallback,
  filterConfig,
  Icon,
  IconType,
  DataTable,
  KubeStatusIndicator,
  LoadingPage,
  statusSortHelper,
  theme,
} from '@weaveworks/weave-gitops';
import { Condition } from '@weaveworks/weave-gitops/ui/lib/api/core/types.pb';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import React, { FC, useCallback, useEffect, useMemo, useState } from 'react';
import { Link, useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { ClusterNamespacedName } from '../../cluster-services/cluster_services.pb';
import useClusters from '../../hooks/clusters';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';
import { useListConfig } from '../../hooks/versions';
import { GitopsClusterEnriched, PRDefaults } from '../../types/custom';
import { useCallbackState } from '../../utils/callback-state';
import {
  EKSDefault,
  GKEDefault,
  Kind,
  Kubernetes,
  Vsphere,
  LiquidMetal,
  Rancher,
  Openshift,
  OtherOnprem,
} from '../../utils/icons';
import { contentCss, ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { TableWrapper, Tooltip } from '../Shared';
import { ConnectClusterDialog } from './ConnectInfoBox';
import { DashboardsList } from './DashboardsList';
import { DeleteClusterDialog } from './Delete';
import { getCreateRequestAnnotation } from './Form/utils';
import { openLinkHandler } from '../../utils/link-checker';

interface Size {
  size?: 'small';
}
const ActionsWrapper = styled.div<Size>`
  display: flex;
`;

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
  max-width: calc(100vw - 220px);
`;

const LoadingWrapper = styled.div`
  ${contentCss};
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

const getClusterTypeIcon = (clusterType?: string): ReactIcon => {
  if (clusterType === 'DockerCluster') {
    return Kind;
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

interface FormData {
  url: string | null;
  branchName: string;
  pullRequestTitle: string;
  commitMessage: string;
  pullRequestDescription: string;
}

const MCCP: FC<{
  location: { state: { notification: NotificationData[] } };
}> = ({ location }) => {
  const { clusters, isLoading } = useClusters();
  const notification = location.state?.notification;
  const [selectedClusters, setSelectedClusters] = useState<
    ClusterNamespacedName[]
  >([]);
  const { setNotifications } = useNotifications();
  const [openConnectInfo, setOpenConnectInfo] = useState<boolean>(false);
  const [openDeletePR, setOpenDeletePR] = useState<boolean>(false);
  const handleClose = useCallback(() => {
    setOpenDeletePR(false);
    setSelectedClusters([]);
  }, [setOpenDeletePR, setSelectedClusters]);
  const { data, repoLink } = useListConfig();
  const repositoryURL = data?.repositoryURL || '';
  const capiClusters = useMemo(
    () => clusters.filter(cls => cls.capiCluster),
    [clusters],
  );
  let selectedCapiClusters = useMemo(
    () =>
      selectedClusters.filter(({ name, namespace }) =>
        capiClusters.find(c => c.name === name && c.namespace === namespace),
      ),
    [capiClusters, selectedClusters],
  );
  const [random, setRandom] = useState<string>(
    Math.random().toString(36).substring(7),
  );
  const classes = useStyles();

  useEffect(() => {
    if (notification) {
      setNotifications(notification);
    }
  }, [notification, setNotifications]);

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
    url: '',
    pullRequestDescription: '',
  };

  const callbackState = useCallbackState();

  if (callbackState) {
    initialFormData = {
      ...initialFormData,
      ...callbackState.state.formData,
    };
    selectedCapiClusters = [
      ...selectedCapiClusters,
      ...(callbackState.state.selectedCapiClusters || []),
    ];
  }

  const [formData, setFormData] = useState<FormData>(initialFormData);
  const history = useHistory();

  const handleAddCluster = useCallback(() => {
    history.push(`/templates`);
  }, [history]);

  const initialFilterState = {
    ...filterConfig(clusters, 'status', filterByStatusCallback),
    ...filterConfig(clusters, 'namespace'),
    ...filterConfig(clusters, 'name'),
  };

  const handleEditCluster = useCallback(
    (event, c) => {
      history.push(`/clusters/${c.name}/edit`);
    },
    [history],
  );

  useEffect(() => {
    if (!callbackState) {
      const prTitle = `Delete clusters: ${selectedCapiClusters
        .map(c => `${c.namespace}/${c.name}`)
        .join(', ')}`;
      setFormData((prevState: FormData) => ({
        ...prevState,
        url: repositoryURL,
        commitMessage: prTitle,
        pullRequestTitle: prTitle,
        pullRequestDescription: prTitle,
      }));
    }

    if (!callbackState && selectedClusters.length === 0) {
      setOpenDeletePR(false);
    }

    if (callbackState?.state?.selectedCapiClusters?.length > 0) {
      setOpenDeletePR(true);
    }
  }, [
    callbackState,
    selectedCapiClusters,
    capiClusters,
    selectedClusters,
    repositoryURL,
  ]);

  const handleSelectAllClick = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      setSelectedClusters(
        clusters.map(({ name, namespace }: GitopsClusterEnriched) => ({
          name,
          namespace,
        })),
      );
      return;
    }
    setSelectedClusters([]);
  };

  const handleIndividualClick = useCallback(
    (
      { name, namespace }: ClusterNamespacedName,
      event: React.ChangeEvent<HTMLInputElement>,
    ) => {
      if (event.target.checked === true) {
        setSelectedClusters((prevState: ClusterNamespacedName[]) => [
          ...prevState,
          { name, namespace },
        ]);
      } else {
        setSelectedClusters((prevState: ClusterNamespacedName[]) =>
          prevState.filter(
            cls => cls.name !== name && cls.namespace !== namespace,
          ),
        );
      }
    },
    [setSelectedClusters],
  );

  const numSelected = selectedClusters.length;
  const rowCount = clusters.length || 0;

  return (
    <PageTemplate
      documentTitle="Clusters"
      path={[{ label: 'Clusters', url: 'clusters' }]}
    >
      <CallbackStateContextProvider
        callbackState={{
          page: authRedirectPage as PageRoute,
          state: { formData, selectedCapiClusters },
        }}
      >
        <ContentWrapper>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}
          >
            <ActionsWrapper>
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
                title="No CAPI clusters selected"
                placement="top"
                disabled={selectedCapiClusters.length !== 0}
              >
                <div>
                  <Button
                    id="delete-cluster"
                    startIcon={<Icon type={IconType.DeleteIcon} size="base" />}
                    onClick={() => {
                      setNotifications([]);
                      setOpenDeletePR(true);
                    }}
                    color="secondary"
                    disabled={selectedCapiClusters.length === 0}
                  >
                    CREATE A PR TO DELETE CLUSTERS
                  </Button>
                </div>
              </Tooltip>
              {openDeletePR && (
                <DeleteClusterDialog
                  formData={formData}
                  setFormData={setFormData}
                  selectedCapiClusters={selectedCapiClusters}
                  onClose={handleClose}
                  prDefaults={PRdefaults}
                />
              )}
              {openConnectInfo && (
                <ConnectClusterDialog
                  onFinish={() => setOpenConnectInfo(false)}
                />
              )}
              <Button onClick={openLinkHandler(repoLink)}>
                <Icon
                  className={classes.externalIcon}
                  type={IconType.ExternalTab}
                  size="base"
                />
                GO TO OPEN PULL REQUESTS
              </Button>
            </ActionsWrapper>
          </div>
          {!isLoading ? (
            <ClustersTableWrapper id="clusters-list">
              <DataTable
                key={clusters.length}
                filters={initialFilterState}
                rows={clusters}
                fields={[
                  {
                    label: 'checkbox',
                    labelRenderer: () => (
                      <Checkbox
                        indeterminate={
                          numSelected > 0 && numSelected < rowCount
                        }
                        checked={rowCount > 0 && numSelected === rowCount}
                        onChange={handleSelectAllClick}
                        inputProps={{ 'aria-label': 'select all rows' }}
                        style={{
                          color: theme.colors.primary,
                        }}
                      />
                    ),
                    value: ({ name, namespace }: GitopsClusterEnriched) => (
                      <ClusterRowCheckbox
                        name={name}
                        namespace={namespace}
                        onChange={handleIndividualClick}
                        checked={Boolean(
                          selectedClusters.find(
                            cls =>
                              cls.name === name && cls.namespace === namespace,
                          ),
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
                    value: (c: GitopsClusterEnriched) => (
                      <Button
                        id="edit-cluster"
                        startIcon={<EditIcon fontSize="small" />}
                        onClick={event => handleEditCluster(event, c)}
                        disabled={!Boolean(getCreateRequestAnnotation(c))}
                      >
                        EDIT CLUSTER
                      </Button>
                    ),
                  },
                ]}
              />
            </ClustersTableWrapper>
          ) : (
            <LoadingWrapper>
              <LoadingPage />
            </LoadingWrapper>
          )}
        </ContentWrapper>
      </CallbackStateContextProvider>
    </PageTemplate>
  );
};

export default MCCP;
