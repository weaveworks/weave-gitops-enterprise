import React, { FC, useCallback, useEffect, useMemo, useState } from 'react';
import { Link, useHistory } from 'react-router-dom';
import styled from 'styled-components';
import {
  Checkbox,
  createStyles,
  makeStyles,
  withStyles,
} from '@material-ui/core';
import Octicon, { Icon as ReactIcon } from '@primer/octicons-react';
import {
  Button,
  filterByStatusCallback,
  filterConfig,
  Icon,
  IconType,
  DataTable,
  KubeStatusIndicator,
  LoadingPage,
  statusSortHelper,
  theme,
  useListSources,
  GitRepository,
  Kind,
} from '@weaveworks/weave-gitops';
import { Condition } from '@weaveworks/weave-gitops/ui/lib/api/core/types.pb';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import { ClusterNamespacedName } from '../../cluster-services/cluster_services.pb';
import useClusters from '../../hooks/clusters';
import { GitopsClusterEnriched, PRDefaults } from '../../types/custom';
import { useCallbackState } from '../../utils/callback-state';
import {
  EKSDefault,
  GKEDefault,
  KindIcon,
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
import { openLinkHandler } from '../../utils/link-checker';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';
import { EditButton } from '../Templates/Edit/EditButton';
import { useListConfigContext } from '../../contexts/ListConfig';
import CallbackStateContextProvider from '../../contexts/GitAuth/CallbackStateContext';
import _ from 'lodash';
import { Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import {
  getCreateRequestAnnotation,
  getInitialGitRepo,
} from '../Templates/Form/utils';
import { GitRepositoryEnriched } from '../Templates/Form';
import { GitProvider } from '../../api/gitauth/gitauth.pb';
import OpenedPullRequest from './OpenedPullRequest';

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
    ['name'],
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

  const gitRepos = useMemo(
    () => getGitRepos(sources?.result),
    [sources?.result],
  );

  const listConfigContext = useListConfigContext();
  // const repoLink = listConfigContext?.repoLink || '';
  const provider = listConfigContext?.provider;

  const gitReposUrl = useMemo(
    () => gitRepos.map(repo => repo.obj.spec.url),
    [gitRepos],
  );

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
  const classes = useStyles();

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
  const initialGitRepo =
    selectedCapiCluster &&
    (getInitialGitRepo(initialUrl, gitRepos) as GitRepositoryEnriched);
  const history = useHistory();

  const handleAddCluster = useCallback(
    () => history.push(`/templates`),
    [history],
  );

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
    <PageTemplate documentTitle="Clusters" path={[{ label: 'Clusters' }]}>
      <CallbackStateContextProvider
        callbackState={{
          page: authRedirectPage as PageRoute,
          state: { formData, selectedCapiCluster },
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
                    color="secondary"
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
              {/* <Button onClick={openLinkHandler(repoLink)}>
                <Icon
                  className={classes.externalIcon}
                  type={IconType.ExternalTab}
                  size="base"
                />
                GO TO OPEN PULL REQUESTS
              </Button> */}
              <OpenedPullRequest options={gitReposUrl}></OpenedPullRequest>
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
                    value: (c: GitopsClusterEnriched) => (
                      <EditButton resource={c} />
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
