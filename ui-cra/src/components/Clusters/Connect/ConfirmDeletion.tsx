import React, { FC, useState } from 'react';
import DialogContentText from '@material-ui/core/DialogContentText';
import { Button as WButton } from '@weaveworks/weave-gitops';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import { faTrashAlt } from '@fortawesome/free-solid-svg-icons';
import { OnClickAction } from '../../Action';

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
      <OnClickAction
        id="delete-cluster"
        icon={faTrashAlt}
        onClick={handleClickOpen}
        text={title}
        className="danger"
        disabled={clusters.length === 0}
      />
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
          <WButton
            onClick={handleClickRemove}
            color="secondary"
            startIcon={<i className="fas fa-trash" />}
          >
            Remove
          </WButton>
          <WButton variant="text" onClick={handleClose}>
            Cancel
          </WButton>
        </DialogActions>
      </Dialog>
    </div>
  );
};

export default ConfirmDeletion;
