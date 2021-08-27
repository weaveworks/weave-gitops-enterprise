import React, { FC, useCallback, useState } from 'react';
import useClusters from '../../contexts/Clusters';
import { Cluster } from '../../types/kubernetes';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import {
  faArrowUp,
  faPlus,
  faTrashAlt,
} from '@fortawesome/free-solid-svg-icons';
import { Snackbar } from '@material-ui/core';
import { ClustersTable } from './Table';
import { FinishMessage, Tooltip } from '../Shared';
import { ConnectClusterDialog } from './Connect/ConnectDialog';
import { useHistory } from 'react-router-dom';
import useTemplates from '../../contexts/Templates';
import { ContentWrapper } from '../Layout/ContentWrapper';
import styled from 'styled-components';
import { OnClickAction } from '../Action';
import theme from 'weaveworks-ui-components/lib/theme';
import { DeleteClusterDialog } from './Create/Delete';

interface Size {
  size?: 'small';
}

const ActionsWrapper = styled.div<Size>`
  display: flex;
  & > * {
    margin-right: ${theme.spacing.medium};
  }
`;

const Title = styled.div`
  font-size: ${theme.fontSizes.large};
  font-weight: 600;
  padding-bottom: ${theme.spacing.medium};
  color: ${theme.colors.gray600};
`;

const MCCP: FC = () => {
  const {
    clusters,
    count,
    disabled,
    handleRequestSort,
    handleSetPageParams,
    order,
    orderBy,
    selectedClusters,
  } = useClusters();

  const [clusterToEdit, setClusterToEdit] = useState<Cluster | null>(null);
  const [finishMessage, setFinishStatus] = useState<FinishMessage | null>(null);
  const [openDeletePR, setOpenDeletePR] = useState<boolean>(false);

  const NEW_CLUSTER = {
    name: '',
    token: '',
  };

  const history = useHistory();
  const { activeTemplate } = useTemplates();
  // const { activeTemplate } = useTemplates();

  const handleAddCluster = useCallback(() => {
    if (activeTemplate === null) {
      history.push('/clusters/templates');
      return null;
    }
    history.push(`/clusters/templates/${activeTemplate.name}/create`);
  }, [activeTemplate, history]);

  const capiClusters = clusters.filter(cls => cls.capiCluster);

  const selectedCapiClusters = selectedClusters.filter(cls =>
    capiClusters.find(c => c.name === cls),
  );

  console.log(openDeletePR);
  console.log(selectedCapiClusters);

  return (
    <PageTemplate documentTitle="WeGo · Clusters">
      <SectionHeader
        className="count-header"
        path={[{ label: 'Clusters', url: 'clusters', count }]}
      />
      <ContentWrapper>
        <Title>Connected clusters dashboard</Title>
        <ActionsWrapper>
          <OnClickAction
            id="create-cluster"
            icon={faPlus}
            onClick={handleAddCluster}
            text="CREATE A CLUSTER"
          />
          <OnClickAction
            id="connect-cluster"
            icon={faArrowUp}
            onClick={() => setClusterToEdit(NEW_CLUSTER)}
            text="CONNECT A CLUSTER"
          />
          <Tooltip
            title="No CAPI clusters selected"
            placement="top"
            disabled={selectedCapiClusters.length !== 0}
          >
            <div>
              <OnClickAction
                className="danger"
                id="delete-cluster"
                icon={faTrashAlt}
                onClick={() => {
                  setOpenDeletePR(true);
                  // setNotifications([]);
                }}
                text="CREATE A PR TO DELETE CLUSTERS"
                disabled={selectedCapiClusters.length === 0}
              />
            </div>
          </Tooltip>
          {openDeletePR && (
            <DeleteClusterDialog
              selectedCapiClusters={selectedCapiClusters}
              setOpenDeletePR={setOpenDeletePR}
            />
          )}
        </ActionsWrapper>
        {clusterToEdit && (
          <ConnectClusterDialog
            cluster={clusterToEdit}
            onFinish={status => {
              setClusterToEdit(null);
              setFinishStatus(status);
            }}
          />
        )}
        <ClustersTable
          onEdit={cluster => {
            setClusterToEdit(cluster);
          }}
          order={order}
          orderBy={orderBy}
          onSortChange={handleRequestSort}
          onSelectPageParams={handleSetPageParams}
          filteredClusters={clusters}
          count={count}
          disabled={disabled}
        />
        <Snackbar
          autoHideDuration={5000}
          open={Boolean(finishMessage?.message)}
          message={finishMessage?.message}
          onClose={() => setFinishStatus(null)}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default MCCP;
