import React, { FC, useState } from 'react';
import styled from 'styled-components';
import DialogContentText from '@material-ui/core/DialogContentText';
import { Button as WButton } from 'weaveworks-ui-components';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import { faTrashAlt } from '@fortawesome/free-solid-svg-icons';
import { OnClickAction } from '../../Action';

const ButtonText = styled.span`
  margin: 0 4px;
`;

const ConfirmDeletion: FC<{
  clusters: number[] | string[];
  onClickRemove: Function;
  title: string;
  formData?: any;
}> = ({ clusters, title, onClickRemove, formData }) => {
  const [open, setOpen] = useState(false);

  const handleClickOpen = () => setOpen(true);
  const handleClose = () => setOpen(false);
  const handleClickRemove = () => {
    onClickRemove({ clusters, formData });
    setOpen(false);
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
          <WButton onClick={handleClickRemove} danger>
            <ButtonText>Remove</ButtonText> <i className="fas fa-trash" />
          </WButton>
          <Button onClick={handleClose} color="primary" autoFocus>
            Cancel
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
};

export default ConfirmDeletion;
