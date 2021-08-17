import React, { ChangeEvent, FC, useCallback, useState } from 'react';
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
  }),
);

export const DeleteClusterDialog: FC<Props> = ({ clusters, onClose }) => {
  const classes = useStyles();
  const random = Math.random().toString(36).substring(7);
  const [branchName, setBranchName] = useState<string>(
    `delete-clusters-branch-${random}`,
  );
  const [pullRequestTitle, setPullRequestTitle] = useState<string>(
    'Deletes capi cluster(s)',
  );
  const [commitMessage, setCommitMessage] = useState<string>(
    'Deletes capi cluster(s)',
  );
  const [pullRequestDescription, setPullRequestDescription] = useState<string>(
    `Delete clusters: ${clusters.map(c => c).join(', ')}`,
  );

  const { deleteCreatedClusters, creatingPR } = useClusters();

  const handleChangeBranchName = useCallback(
    (event: ChangeEvent<HTMLInputElement>) => setBranchName(event.target.value),
    [],
  );

  const handleChangePullRequestTitle = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setPullRequestTitle(event.target.value),
    [],
  );

  const handleChangeCommitMessage = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setCommitMessage(event.target.value),
    [],
  );

  const handleChangePRDescription = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setPullRequestDescription(event.target.value),
    [],
  );

  const handleClickRemove = () =>
    deleteCreatedClusters({
      clusterNames: clusters,
      headBranch: branchName,
      title: pullRequestTitle,
      commitMessage,
      description: pullRequestDescription,
    });

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
              <Input
                label="Create branch"
                placeholder={branchName}
                onChange={handleChangeBranchName}
              />
              <Input
                label="Pull request title"
                placeholder={pullRequestTitle}
                onChange={handleChangePullRequestTitle}
              />
              <Input
                label="Commit message"
                placeholder={commitMessage}
                onChange={handleChangeCommitMessage}
              />
              <Input
                label="Pull request description"
                placeholder={pullRequestDescription}
                onChange={handleChangePRDescription}
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
