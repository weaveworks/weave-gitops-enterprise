import React, { FC, useCallback, useEffect, useMemo, useState } from 'react';
import { ThemeProvider } from '@material-ui/core/styles';
import useClusters from '../../contexts/Clusters';
import useNotifications from '../../contexts/Notifications';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ClustersTable } from './Table';
import { Tooltip } from '../Shared';
import { ConnectClusterDialog } from './Connect/ConnectDialog';
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
} from '@weaveworks/weave-gitops';
import { DeleteClusterDialog } from './Delete';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import useVersions from '../../contexts/Versions';
import { localEEMuiTheme } from '../../muiTheme';
import { GitopsCluster } from '../../capi-server/capi_server.pb';

interface Size {
  size?: 'small';
}

const ActionsWrapper = styled.div<Size>`
  display: flex;
  & > * {
    margin-right: ${theme.spacing.medium};
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
  } = useClusters();
  const { setNotifications } = useNotifications();
  const [clusterToEdit, setClusterToEdit] = useState<GitopsCluster | null>(
    null,
  );
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

  const NEW_CLUSTER = {
    name: '',
    token: '',
  };

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
                onClick={() => setClusterToEdit(NEW_CLUSTER)}
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
            {clusterToEdit && (
              <ConnectClusterDialog
                cluster={clusterToEdit}
                onFinish={() => setClusterToEdit(null)}
              />
            )}
            <ClustersTable
              onEdit={cluster => setClusterToEdit(cluster)}
              order={order}
              orderBy={orderBy}
              onSortChange={handleRequestSort}
              filteredClusters={clusters}
              count={count}
              disabled={disabled}
            />
          </ContentWrapper>
        </CallbackStateContextProvider>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default MCCP;
