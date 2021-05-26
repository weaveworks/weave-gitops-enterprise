import React, { FC, useState } from 'react';
import styled from 'styled-components';
import { FormState, SetFormState } from '../../../types/form';
import Box from '@material-ui/core/Box';
import DialogContentText from '@material-ui/core/DialogContentText';
import Typography from '@material-ui/core/Typography';
import { Button as WButton, CircularProgress } from 'weaveworks-ui-components';

import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import { Poll } from '../../../utils/poll';
import { asMilliseconds } from '../../../utils/time';
import { Cluster } from '../../../types/kubernetes';
import { request } from '../../../utils/request';
import { Code, HandleFinish, Status } from '../../Shared';

const Container = styled.div`
  margin-right: 16px;
  margin-left: 16px;
`;

interface ResponsesById {
  getCluster?: Cluster;
}

const ButtonText = styled.span`
  margin: 0 4px;
`;

const ConfirmDeletion: FC<{ clusterId: number; onFinish: HandleFinish }> = ({
  clusterId,
  onFinish,
}) => {
  const [open, setOpen] = useState(false);
  const [submitting, setSubmitting] = useState<boolean>(false);
  const [errors, setError] = useState<string>();

  const handleClickOpen = () => {
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
  };

  const handleClickRemove = () => {
    setSubmitting(true);
    request('DELETE', `/gitops/api/clusters/${clusterId}`)
      .then(() => {
        setSubmitting(false);
        setOpen(false);
        onFinish({
          success: true,
          message: 'Cluster successfully removed from the MCCP',
        });
      })
      .catch(({ message }) => {
        setError(message);
        setSubmitting(false);
      });
  };

  return (
    <div>
      <WButton onClick={handleClickOpen} danger>
        <ButtonText>Remove cluster from the MCCP</ButtonText>{' '}
        <i className="fas fa-trash" />
      </WButton>
      <Dialog
        open={open}
        onClose={handleClose}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
        id="confirm-disconnect-cluster-dialog"
      >
        <DialogTitle id="alert-dialog-title">
          Remove cluster from the MCCP
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">
            Are you sure you want to remove this cluster from the MCCP?
            {errors && <Typography color="error">{errors}</Typography>}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <WButton disabled={submitting} onClick={handleClickRemove} danger>
            <ButtonText>Remove</ButtonText> <i className="fas fa-trash" />
          </WButton>
          <Button
            disabled={submitting}
            onClick={handleClose}
            color="primary"
            autoFocus
          >
            Cancel
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
};

export const ClusterDisconnectionInstructions: FC<{
  formState: FormState;
  setFormState: SetFormState;
  onFinish: HandleFinish;
}> = ({ formState, onFinish }) => {
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
      <ConfirmDeletion clusterId={formState.cluster.id} onFinish={onFinish} />
    </Container>
  );
};
