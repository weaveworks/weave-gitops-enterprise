import React, { FC, useCallback, useEffect, useMemo, useState } from 'react';
import { ThemeProvider } from '@material-ui/core/styles';
import useClusters from '../../contexts/Clusters';
import useNotifications from '../../contexts/Notifications';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { Tooltip } from '../Shared';
import { ConnectClusterDialog } from './ConnectInfoBox';
import { Link, useHistory } from 'react-router-dom';
import useTemplates from '../../contexts/Templates';
import { contentCss, ContentWrapper, Title } from '../Layout/ContentWrapper';
import styled from 'styled-components';
import {
  Button,
  theme,
  CallbackStateContextProvider,
  getCallbackState,
  Icon,
  IconType,
  // filterConfigForString,
  FilterableTable,
  filterByStatusCallback,
  filterConfig,
  LoadingPage,
  KubeStatusIndicator,
  SortType,
  statusSortHelper,
  applicationsClient,
} from '@weaveworks/weave-gitops';
import { DeleteClusterDialog } from './Delete';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import useVersions from '../../contexts/Versions';
import { localEEMuiTheme } from '../../muiTheme';
import { Checkbox, withStyles } from '@material-ui/core';
import { GitopsClusterEnriched } from '../../types/custom';
import { DashboardsList } from './DashboardsList';

interface Size {
  size?: 'small';
}

const ActionsWrapper = styled.div<Size>`
  display: flex;
  & > * {
    margin-right: ${theme.spacing.medium};
  }
`;

const TableWrapper = styled.div`
  margin-top: ${theme.spacing.medium};
  div[class*='FilterDialog__SlideContainer'],
  div[class*='SearchField'] {
    overflow: hidden;
  }
  div[class*='FilterDialog'] {
    .Mui-checked {
      color: ${theme.colors.primary};
    }
  }
  thead {
    th:first-of-type {
      padding: ${theme.spacing.base};
    }
  }
  td:first-of-type {
    text-overflow: clip;
    width: 25px;
  }
  td:nth-child(2) {
    width: 650px;
  }
  a {
    color: ${theme.colors.primary};
  }
  max-width: calc(100vw - 220px);
`;

const LoadingWrapper = styled.div`
  ${contentCss};
`;

const random = Math.random().toString(36).substring(7);

export const PRdefaults = {
  branchName: `delete-clusters-branch-${random}`,
  pullRequestTitle: 'Deletes capi cluster(s)',
  commitMessage: 'Deletes capi cluster(s)',
};

const MCCP: FC = () => {
  const { clusters, isLoading, count, selectedClusters, setSelectedClusters } =
    useClusters();
  const { setNotifications } = useNotifications();
  const [openConnectInfo, setOpenConnectInfo] = useState<boolean>(false);
  const [openDeletePR, setOpenDeletePR] = useState<boolean>(false);
  const { repositoryURL } = useVersions();
  const [repoLink, setRepoLink] = useState<string>('');
  const capiClusters = useMemo(
    () => clusters.filter(cls => cls.capiCluster),
    [clusters],
  );
  let selectedCapiClusters = useMemo(
    () =>
      selectedClusters.filter(cls => capiClusters.find(c => c.name === cls)),
    [capiClusters, selectedClusters],
  );

  const authRedirectPage = `/clusters`;

  interface FormData {
    url: string | null;
    branchName: string;
    pullRequestTitle: string;
    commitMessage: string;
    pullRequestDescription: string;
  }

  let initialFormData = {
    ...PRdefaults,
    url: '',
    pullRequestDescription: '',
  };

  const callbackState = getCallbackState();

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
  const { activeTemplate } = useTemplates();

  const handleAddCluster = useCallback(() => {
    if (activeTemplate === null) {
      history.push('/clusters/templates');
      return null;
    }
    history.push(`/clusters/templates/${activeTemplate.name}/create`);
  }, [activeTemplate, history]);

  const initialFilterState = {
    ...filterConfig(clusters, 'status', filterByStatusCallback),
    ...filterConfig(clusters, 'namespace'),
  };

  useEffect(() => {
    if (!callbackState) {
      setFormData((prevState: FormData) => ({
        ...prevState,
        url: repositoryURL,
        pullRequestDescription: `Delete clusters: ${selectedCapiClusters
          .map(c => c)
          .join(', ')}`,
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
      const newSelected =
        clusters.map((cluster: GitopsClusterEnriched) => cluster.name || '') ||
        [];
      setSelectedClusters(newSelected);
      return;
    }
    setSelectedClusters([]);
  };

  const handleIndividualClick = (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    if (event.target.checked === true) {
      setSelectedClusters((prevState: string[]) => [
        ...prevState,
        event.target.name,
      ]);
    } else {
      setSelectedClusters((prevState: string[]) =>
        prevState.filter(cls => event.target.name !== cls),
      );
    }
  };

  const numSelected = selectedClusters.length;
  const rowCount = clusters.length || 0;

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

  useEffect(() => {
    repositoryURL &&
      applicationsClient.ParseRepoURL({ url: repositoryURL }).then(res => {
        if (res.provider === 'GitHub') {
          setRepoLink(repositoryURL + `/pulls`);
        } else if (res.provider === 'GitLab') {
          setRepoLink(repositoryURL + `/-/merge_requests`);
        }
      });
  }, [repositoryURL]);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Clusters">
        <CallbackStateContextProvider
          callbackState={{
            page: authRedirectPage as PageRoute,
            state: { formData, selectedCapiClusters },
          }}
        >
          <SectionHeader
            className="count-header"
            path={[{ label: 'Clusters', url: 'clusters', count }]}
          />
          <ContentWrapper>
            <Title>Connected clusters dashboard</Title>
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
                  startIcon={
                    <Icon type={IconType.ArrowUpwardIcon} size="base" />
                  }
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
                      startIcon={
                        <Icon type={IconType.DeleteIcon} size="base" />
                      }
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
                    setOpenDeletePR={setOpenDeletePR}
                  />
                )}
                {openConnectInfo && (
                  <ConnectClusterDialog
                    onFinish={() => setOpenConnectInfo(false)}
                  />
                )}
              </ActionsWrapper>
              <a
                style={{
                  color: theme.colors.primary,
                  padding: theme.spacing.small,
                }}
                href={repoLink}
                target="_blank"
                rel="noopener noreferrer"
              >
                View open Pull Requests
              </a>
            </div>
            {!isLoading ? (
              <TableWrapper id="clusters-list">
                <FilterableTable
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
                      value: (c: GitopsClusterEnriched) => (
                        <IndividualCheckbox
                          checked={
                            selectedClusters.indexOf(c.name ? c.name : '') !==
                            -1
                          }
                          onChange={handleIndividualClick}
                          name={c.name}
                        />
                      ),
                      maxWidth: 25,
                    },
                    {
                      label: 'Name',
                      value: (c: GitopsClusterEnriched) => (
                        <Link
                          to={`/cluster?clusterName=${c.name}`}
                          color={theme.colors.primary}
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
                      value: (c: GitopsClusterEnriched) =>
                        c.capiClusterRef ? 'capi' : 'other',
                    },
                    {
                      label: 'Namespace',
                      value: 'namespace',
                    },
                    {
                      label: 'Status',
                      value: (c: GitopsClusterEnriched) =>
                        c.conditions && c.conditions.length > 0 ? (
                          <KubeStatusIndicator
                            short
                            conditions={c.conditions}
                          />
                        ) : null,
                      sortType: SortType.number,
                      sortValue: statusSortHelper,
                    },
                  ]}
                />
              </TableWrapper>
            ) : (
              <LoadingWrapper>
                <LoadingPage />
              </LoadingWrapper>
            )}
          </ContentWrapper>
        </CallbackStateContextProvider>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default MCCP;
