import React, {
  ChangeEvent,
  FC,
  useCallback,
  useEffect,
  useState,
  Dispatch,
} from 'react';
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
import useNotifications from '../../../contexts/Notifications';
import { faTrashAlt } from '@fortawesome/free-solid-svg-icons';
import { OnClickAction } from '../../Action';
import { Input } from '../../../utils/form';
import { Loader } from '../../Loader';

interface Props {
  selectedCapiClusters: string[];
  setOpenDeletePR: Dispatch<React.SetStateAction<boolean>>;
}

const useStyles = makeStyles(() =>
  createStyles({
    dialog: {
      backgroundColor: theme.colors.gray50,
    },
  }),
);

export const DeleteClusterDialog: FC<Props> = ({
  selectedCapiClusters,
  setOpenDeletePR,
}) => {
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
    `Delete clusters: ${selectedCapiClusters.map(c => c).join(', ')}`,
  );

  const { deleteCreatedClusters, creatingPR, setSelectedClusters } =
    useClusters();
  const { notifications, setNotifications } = useNotifications();

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
      clusterNames: selectedCapiClusters,
      headBranch: branchName,
      title: pullRequestTitle,
      commitMessage,
      description: pullRequestDescription,
    });

  useEffect(() => {
    if (
      notifications.length > 0 &&
      notifications[notifications.length - 1].variant !== 'danger'
    ) {
      setOpenDeletePR(false);
      setSelectedClusters([]);
      // setNotifications([]);
    }
    // setOpenDeletePR(true);
    // return setNotifications([]);
  }, [notifications, setOpenDeletePR, setSelectedClusters]);

  return (
    <Dialog
      open
      maxWidth="md"
      fullWidth
      onClose={() => {
        setOpenDeletePR(false);
        setSelectedClusters([]);
      }}
    >
      <div id="delete-popup" className={classes.dialog}>
        <DialogTitle disableTypography>
          <Typography variant="h5">Create PR to remove clusters</Typography>
          {setOpenDeletePR ? (
            <CloseIconButton
              onClick={() => {
                setOpenDeletePR(false);
                setSelectedClusters([]);
              }}
            />
          ) : null}
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
                disabled={selectedCapiClusters.length === 0}
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
