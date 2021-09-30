import React, { FC } from 'react';
import styled from 'styled-components';
import { FormState, SetFormState } from '../../../types/form';
import DialogContentText from '@material-ui/core/DialogContentText';
import Typography from '@material-ui/core/Typography';
import { Poll } from '../../../utils/poll';
import { asMilliseconds } from '../../../utils/time';
import { Cluster } from '../../../types/kubernetes';
import { Code, statusBox } from '../../Shared';
import ConfirmDeletion from './ConfirmDeletion';
import useClusters from '../../../contexts/Clusters';
import { Loader } from '../../Loader';

const Container = styled.div`
  margin-right: 16px;
  margin-left: 16px;
`;

interface ResponsesById {
  getCluster?: Cluster;
}

export const ClusterDisconnectionInstructions: FC<{
  formState: FormState;
  setFormState: SetFormState;
  connecting: boolean;
}> = ({ formState, connecting }) => {
  const { deleteConnectedClusters } = useClusters();

  if (!formState.cluster.id) {
    return (
      <Typography color="error">No Cluster ID, not created yet?</Typography>
    );
  }
  const getCluster = `/gitops/api/clusters/${formState.cluster.id}`;
  const { protocol, host } = window.location;
  const yamlUrl = `${protocol}//${host}/gitops/api/agent.yaml?token=${formState.cluster.token}`;

  return (
    <Container>
      <DialogContentText>
        To disconnect {formState.cluster.name} first remove the agent by
        running:
      </DialogContentText>
      <Code id="instructions">kubectl delete -f {yamlUrl}</Code>
      <Poll<ResponsesById>
        intervalMs={asMilliseconds('5s')}
        queriesById={{ getCluster }}
      >
        {({ responsesById: { getCluster: cluster } }) => {
          if (!cluster) {
            return <Loader />;
          }
          return statusBox(cluster, connecting);
        }}
      </Poll>
      <DialogContentText>
        Finally, to remove the cluster and data that has been collected in the
        MCCP database click the button below. This operation will not destroy or
        modify the cluster itself.
      </DialogContentText>
      <ConfirmDeletion
        clusters={[formState.cluster.id]}
        onClickRemove={deleteConnectedClusters}
        title="Remove cluster from the MCCP"
      />
    </Container>
  );
};
