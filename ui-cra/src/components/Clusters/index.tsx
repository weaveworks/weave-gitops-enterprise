import React, { FC, useCallback, useState } from 'react';
import useClusters from '../../contexts/Clusters';
import useNotifications from '../../contexts/Notifications';
import { Cluster } from '../../types/kubernetes';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import {
  faArrowUp,
  faPlus,
  faTrashAlt,
} from '@fortawesome/free-solid-svg-icons';
import { ClustersTable } from './Table';
import { Tooltip } from '../Shared';
import { ConnectClusterDialog } from './Connect/ConnectDialog';
import { useHistory } from 'react-router-dom';
import useTemplates from '../../contexts/Templates';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
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
  const { setNotifications } = useNotifications();
  const [clusterToEdit, setClusterToEdit] = useState<Cluster | null>(null);
  const [openDeletePR, setOpenDeletePR] = useState<boolean>(false);

  const NEW_CLUSTER = {
    name: '',
    token: '',
  };

  const history = useHistory();
  const { activeTemplate } = useTemplates();

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
                  setNotifications([]);
                  setOpenDeletePR(true);
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
            onFinish={() => setClusterToEdit(null)}
          />
        )}
        <ClustersTable
          onEdit={cluster => setClusterToEdit(cluster)}
          order={order}
          orderBy={orderBy}
          onSortChange={handleRequestSort}
          onSelectPageParams={handleSetPageParams}
          filteredClusters={clusters}
          count={count}
          disabled={disabled}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default MCCP;
