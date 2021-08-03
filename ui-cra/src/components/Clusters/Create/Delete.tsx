import React, { FC, useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogTitle,
  Typography,
} from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import { createStyles } from '@material-ui/styles';
import theme from 'weaveworks-ui-components/lib/theme';
import { CloseIconButton } from '../../../assets/img/close-icon-button';
import { HandleFinish } from '../../Shared';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import DialogContentText from '@material-ui/core/DialogContentText';
import ConfirmDeletion from '../Connect/ConfirmDeletion';
import useClusters from '../../../contexts/Clusters';

interface Props {
  clusters: string[];
  onFinish: HandleFinish;
}

const useStyles = makeStyles(() =>
  createStyles({
    dialog: {
      backgroundColor: theme.colors.gray50,
    },
    textarea: {
      width: '100%',
      padding: weaveTheme.spacing.xs,
      marginBottom: weaveTheme.spacing.xs,
      border: '1px solid #E5E5E5',
    },
  }),
);

export const DeleteClusterDialog: FC<Props> = ({ clusters, onFinish }) => {
  const onSubmit = (event: any) => {
    event.preventDefault();
    event.stopPropagation();
    const [PRDescription] = event.currentTarget.elements;
    setFormData(PRDescription.value);
  };
  const classes = useStyles();
  const [formData, setFormData] = useState();
  const { deleteCreatedClusters } = useClusters();

  return (
    <Dialog
      maxWidth="md"
      fullWidth
      onClose={() => onFinish({ success: true, message: '' })}
      open
    >
      <div id="delete-popup" className={classes.dialog}>
        <DialogTitle disableTypography>
          <Typography variant="h5">Create PR to remove clusters</Typography>
          {onFinish ? (
            <CloseIconButton
              onClick={() => onFinish({ success: true, message: '' })}
            />
          ) : null}
        </DialogTitle>
        <DialogContent>
          <DialogContentText>Add a description for your PR:</DialogContentText>
          <form onSubmit={onSubmit}>
            <textarea className={classes.textarea} rows={2} />
            <ConfirmDeletion
              formData={formData}
              clusters={clusters}
              onClickRemove={deleteCreatedClusters}
              title="Remove clusters from the MCCP"
            />
          </form>
        </DialogContent>
      </div>
    </Dialog>
  );
};
