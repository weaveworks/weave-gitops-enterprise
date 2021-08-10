import React, { ChangeEvent, FC, useState } from 'react';
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
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import DialogContentText from '@material-ui/core/DialogContentText';
import useClusters from '../../../contexts/Clusters';
import { faTrashAlt } from '@fortawesome/free-solid-svg-icons';
import { OnClickAction } from '../../Action';
import { Input } from '../../../utils/form';
import { Loader } from '../../Loader';

interface Props {
  clusters: string[];
  onClose: () => void;
}

const useStyles = makeStyles(() =>
  createStyles({
    dialog: {
      backgroundColor: theme.colors.gray50,
    },
    input: {
      marginBottom: weaveTheme.spacing.xs,
    },
  }),
);

export const DeleteClusterDialog: FC<Props> = ({ clusters, onClose }) => {
  const classes = useStyles();
  const [PRDescription, setPRDescription] = useState<string>();
  const { deleteCreatedClusters, creatingPR } = useClusters();

  const handleChangePRDescription = (event: ChangeEvent<HTMLInputElement>) =>
    setPRDescription(event.target.value);

  const handleClickRemove = () =>
    deleteCreatedClusters(clusters, PRDescription);

  return (
    <Dialog open maxWidth="md" fullWidth onClose={() => onClose()}>
      <div id="delete-popup" className={classes.dialog}>
        <DialogTitle disableTypography>
          <Typography variant="h5">Create PR to remove clusters</Typography>
          {onClose ? <CloseIconButton onClick={() => onClose()} /> : null}
        </DialogTitle>
        <DialogContent>
          {!creatingPR ? (
            <>
              <DialogContentText>
                Add a description for your PR:
              </DialogContentText>
              <Input
                onChange={handleChangePRDescription}
                className={classes.input}
                multiline
                rows={4}
              />
              <OnClickAction
                id="delete-cluster"
                icon={faTrashAlt}
                onClick={handleClickRemove}
                text="Remove clusters from the MCCP"
                className="danger"
                disabled={clusters.length === 0}
              />
            </>
          ) : (
            <Loader />
          )}
        </DialogContent>
      </div>
    </Dialog>
  );
};
