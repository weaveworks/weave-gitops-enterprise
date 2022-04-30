import React, { FC } from 'react';
import {
  Dialog,
  DialogContent,
  DialogTitle,
  Typography,
} from '@material-ui/core';
import { CloseIconButton } from '../../assets/img/close-icon-button';

interface Props {
  onFinish: () => void;
}
export const ConnectClusterDialog: FC<Props> = ({ onFinish }) => {
  return (
    <Dialog maxWidth="md" fullWidth onClose={() => onFinish()} open>
      <div id="connection-popup">
        <DialogTitle disableTypography>
          <Typography variant="h5">Connect a cluster</Typography>
          {onFinish ? <CloseIconButton onClick={() => onFinish()} /> : null}
        </DialogTitle>
        <DialogContent>Here is where we explain how to connect</DialogContent>
      </div>
    </Dialog>
  );
};
