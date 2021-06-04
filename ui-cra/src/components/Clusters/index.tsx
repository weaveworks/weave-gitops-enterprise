import React, { FC, useState } from 'react';
import useClusters from '../../contexts/Clusters';
import { Cluster } from '../../types/kubernetes';
import { PageTemplate } from '../Layout/PageTemplate';
import { OnClickAction, SectionHeader } from '../Layout/SectionHeader';
import { faPlug, faPlus } from '@fortawesome/free-solid-svg-icons';
import { Snackbar } from '@material-ui/core';
import { ClustersTable } from './Table';
import { FinishMessage } from '../Shared';
import { ConnectClusterDialog } from './Connect/ConnectDialog';
import { useHistory } from 'react-router-dom';
import useTemplates from '../../contexts/Templates';

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

  const handleAddCluster = () => {
    if (activeTemplate === null) {
      history.push('/templates');
      return null;
    }
    history.push(`/templates/${activeTemplate.name}/create`);
  };

  return (
    <PageTemplate documentTitle="WeGo · Clusters">
      <span id="count-header">
        <SectionHeader
          actions={[
            <OnClickAction
              id="create-cluster"
              icon={faPlus}
              onClick={handleAddCluster}
              text="Create a cluster with template"
            />,
            <OnClickAction
              id="connect-cluster"
              icon={faPlug}
              onClick={() => setClusterToEdit(NEW_CLUSTER)}
              text="Connect a cluster"
            />,
          ]}
          path={[{ label: 'Clusters', url: 'clusters', count }]}
        />
      </span>
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
    </PageTemplate>
  );
};

export default MCCP;
