import React, { FC } from 'react';
import {
  Dialog,
  DialogContent,
  DialogTitle,
  Typography,
} from '@material-ui/core';
import { ConnectClusterWizard } from './ConnectWizard';
import { Cluster } from '../../../types/kubernetes';
import { CloseIconButton } from '../../../assets/img/close-icon-button';

interface Props {
  cluster: Cluster;
  onFinish: () => void;
}
export const ConnectClusterDialog: FC<Props> = ({ cluster, onFinish }) => {
  return (
    <Dialog maxWidth="md" fullWidth onClose={() => onFinish()} open>
      <div id="connection-popup">
        <DialogTitle disableTypography>
          <Typography variant="h5">
            {cluster.id ? 'Configure cluster' : 'Connect a cluster'}
          </Typography>
          {onFinish ? <CloseIconButton onClick={() => onFinish()} /> : null}
        </DialogTitle>
        <DialogContent>
          <ConnectClusterWizard
            connecting={!cluster.id}
            cluster={cluster}
            onFinish={onFinish}
          />
        </DialogContent>
      </div>
    </Dialog>
  );
};
