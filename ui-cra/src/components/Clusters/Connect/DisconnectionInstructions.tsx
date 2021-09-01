import React, { FC } from 'react';
import styled from 'styled-components';
import { FormState, SetFormState } from '../../../types/form';
import Box from '@material-ui/core/Box';
import DialogContentText from '@material-ui/core/DialogContentText';
import Typography from '@material-ui/core/Typography';
import { CircularProgress } from 'weaveworks-ui-components';
import { Poll } from '../../../utils/poll';
import { asMilliseconds } from '../../../utils/time';
import { Cluster } from '../../../types/kubernetes';
import { Code, HandleFinish, Status } from '../../Shared';
import ConfirmDeletion from './ConfirmDeletion';
import useClusters from '../../../contexts/Clusters';

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
  onFinish: HandleFinish;
}> = ({ formState, onFinish }) => {
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
            return <CircularProgress size="small" />;
          }
          return (
            <Box lineHeight="24px" display="flex" alignItems="center" my={2}>
              <Box color="text.secondary" mr={1}>
                Cluster status
              </Box>
              <Status updatedAt={cluster.updatedAt} status={cluster.status} />
            </Box>
          );
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
        onFinish={onFinish}
      />
    </Container>
  );
};
