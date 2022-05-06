import React, { FC, useCallback, useEffect, useMemo, useState } from 'react';
import { ThemeProvider } from '@material-ui/core/styles';
import useClusters, { GitopsClusterEnriched } from '../../contexts/Clusters';
import useNotifications from '../../contexts/Notifications';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
// import { ClustersTable } from './Table';
import { Tooltip } from '../Shared';
import { ConnectClusterDialog } from './ConnectInfoBox';
import { useHistory } from 'react-router-dom';
import useTemplates from '../../contexts/Templates';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import styled from 'styled-components';
import {
  Button,
  theme,
  CallbackStateContextProvider,
  getCallbackState,
  Icon,
  IconType,
  filterConfigForString,
  FilterableTable,
  filterConfigForStatus,
} from '@weaveworks/weave-gitops';
import { DeleteClusterDialog } from './Delete';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import useVersions from '../../contexts/Versions';
import { localEEMuiTheme } from '../../muiTheme';
import { Checkbox, withStyles } from '@material-ui/core';

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
  div[class*='FilterDialog__SlideContainer'],
  div[class*='SearchField'] {
    overflow: hidden;
  }
`;

const random = Math.random().toString(36).substring(7);

export const PRdefaults = {
  branchName: `delete-clusters-branch-${random}`,
  pullRequestTitle: 'Deletes capi cluster(s)',
  commitMessage: 'Deletes capi cluster(s)',
};

const MCCP: FC = () => {
  const {
    clusters,
    count,
    disabled,
    handleRequestSort,
    order,
    orderBy,
    selectedClusters,
    setSelectedClusters,
  } = useClusters();
  const { setNotifications } = useNotifications();
  const [openConnectInfo, setOpenConnectInfo] = useState<boolean>(false);
  const [openDeletePR, setOpenDeletePR] = useState<boolean>(false);
  const { repositoryURL } = useVersions();
  const capiClusters = useMemo(
    // @ts-ignore
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
    url: string;
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
    ...filterConfigForString(clusters, 'type'),
    ...filterConfigForString(clusters, 'namespace'),
    ...filterConfigForStatus(clusters),
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

  console.log(clusters);

  // Selection of clusters

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

  const handleClick = (event: React.MouseEvent<unknown>, name?: string) => {
    const selectedIndex = selectedClusters.indexOf(name || '');
    let newSelected: string[] = [];

    if (selectedIndex === -1) {
      newSelected = newSelected.concat(selectedClusters, name || '');
    } else if (selectedIndex === 0) {
      newSelected = newSelected.concat(selectedClusters.slice(1));
    } else if (selectedIndex === selectedClusters.length - 1) {
      newSelected = newSelected.concat(selectedClusters.slice(0, -1));
    } else if (selectedIndex > 0) {
      newSelected = newSelected.concat(
        selectedClusters.slice(0, selectedIndex),
        selectedClusters.slice(selectedIndex + 1),
      );
    }
    setSelectedClusters(newSelected);
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
                  setOpenDeletePR={setOpenDeletePR}
                />
              )}
            </ActionsWrapper>
            {openConnectInfo && (
              <ConnectClusterDialog
                onFinish={() => setOpenConnectInfo(false)}
              />
            )}
            {clusters && (
              <TableWrapper>
                <FilterableTable
                  filters={initialFilterState}
                  rows={clusters}
                  fields={[
                    {
                      label: () => (
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
                            selectedClusters.indexOf(c.name || '') !== -1
                          }
                          // inputProps={{ 'aria-labelledby': labelId }}
                          onClick={(event: any) => handleClick(event, c.name)}
                        />
                      ),
                    },
                    {
                      label: 'Name',
                      value: 'name',
                    },
                    {
                      label: 'Type',
                      value: 'type',
                    },
                    { label: 'Namespace', value: 'namespace' },
                  ]}
                />
              </TableWrapper>
            )}
          </ContentWrapper>
        </CallbackStateContextProvider>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default MCCP;
