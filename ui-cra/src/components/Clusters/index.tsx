import React, { FC, useCallback, useState } from 'react';
import useClusters from '../../contexts/Clusters';
import { Cluster } from '../../types/kubernetes';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { faPlug, faPlus } from '@fortawesome/free-solid-svg-icons';
import { Snackbar } from '@material-ui/core';
import { ClustersTable } from './Table';
import { FinishMessage } from '../Shared';
import { ConnectClusterDialog } from './Connect/ConnectDialog';
import { useHistory } from 'react-router-dom';
import useTemplates from '../../contexts/Templates';
import { ContentWrapper } from '../Layout/ContentWrapper';
import styled from 'styled-components';
import { OnClickAction } from '../Action';
import theme from 'weaveworks-ui-components/lib/theme';

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
    error,
    count,
    loading,
    disabled,
    handleRequestSort,
    handleSetPageParams,
    order,
    orderBy,
  } = useClusters();

  const [clusterToEdit, setClusterToEdit] = useState<Cluster | null>(null);
  const [finishMessage, setFinishStatus] = useState<FinishMessage | null>(null);

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
            icon={faPlug}
            onClick={() => setClusterToEdit(NEW_CLUSTER)}
            text="CONNECT A CLUSTER"
          />
        </ActionsWrapper>
        ,
        {clusterToEdit && (
          <ConnectClusterDialog
            cluster={clusterToEdit}
            onFinish={status => {
              setClusterToEdit(null);
              setFinishStatus(status);
            }}
          />
        )}
        {/* TBD: Do we need to pass down the loading state to Clusters or can we manage this in the Clusters Provider with a loader? */}
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
          error={error}
          disabled={disabled}
          isLoading={loading}
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
