import React, { FC, useState } from 'react';
import DialogContentText from '@material-ui/core/DialogContentText';
import { Button, Icon } from '@weaveworks/weave-gitops';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';

const ConfirmDeletion: FC<{
  clusters: number[] | string[];
  onClickRemove: Function;
  title: string;
  formData?: any;
  onFinish: () => void;
}> = ({ clusters, title, onClickRemove, formData, onFinish }) => {
  const [open, setOpen] = useState(false);

  const handleClickOpen = () => setOpen(true);
  const handleClose = () => setOpen(false);
  const handleClickRemove = () => {
    onClickRemove({ clusters, formData });
    setOpen(false);
    onFinish();
  };

  return (
    <div>
      <Button
        id="delete-cluster"
        color="secondary"
        startIcon={<Icon type="Delete" size="base" />}
        onClick={handleClickOpen}
        disabled={clusters.length === 0}
      >
        REMOVE CLUSTERS FROM THE MCCP
      </Button>
      <Dialog
        open={open}
        onClose={handleClose}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
        id="confirm-disconnect-cluster-dialog"
      >
        <DialogTitle id="alert-dialog-title">
          Delete cluster confirmation
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">
            Are you sure you want to remove this cluster?
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={handleClickRemove}
            color="secondary"
            startIcon={<Icon type="Delete" size="base" />}
          >
            REMOVE
          </Button>
          <Button variant="text" onClick={handleClose}>
            CANCEL
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
};

export default ConfirmDeletion;
