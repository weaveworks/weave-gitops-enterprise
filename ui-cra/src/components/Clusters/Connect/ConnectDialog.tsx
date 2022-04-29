import React, { FC } from 'react';
import {
  Dialog,
  DialogContent,
  DialogTitle,
  Typography,
} from '@material-ui/core';
import { ConnectClusterWizard } from './ConnectWizard';
import { CloseIconButton } from '../../../assets/img/close-icon-button';
import { GitopsCluster } from '../../../capi-server/capi_server.pb';

interface Props {
  cluster: GitopsCluster;
  onFinish: () => void;
}
export const ConnectClusterDialog: FC<Props> = ({ cluster, onFinish }) => {
  return (
    <Dialog maxWidth="md" fullWidth onClose={() => onFinish()} open>
      <div id="connection-popup">
        <DialogTitle disableTypography>
          <Typography variant="h5">
            {cluster.name ? 'Configure cluster' : 'Connect a cluster'}
          </Typography>
          {onFinish ? <CloseIconButton onClick={() => onFinish()} /> : null}
        </DialogTitle>
        <DialogContent>
          <ConnectClusterWizard
            connecting={!cluster.name}
            cluster={cluster}
            onFinish={onFinish}
          />
        </DialogContent>
      </div>
    </Dialog>
  );
};
