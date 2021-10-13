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
import {
  getProviderToken,
  GithubDeviceAuthModal,
} from '@weaveworks/weave-gitops';
import { isUnauthenticated } from '../../../utils/request';

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
  const [showAuthDialog, setShowAuthDialog] = useState(false);
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
    deleteCreatedClusters(
      {
        clusterNames: selectedCapiClusters,
        headBranch: branchName,
        title: pullRequestTitle,
        commitMessage,
        description: pullRequestDescription,
      },
      getProviderToken('github'),
    )
      .then(() =>
        setNotifications([
          {
            message: `PR created successfully`,
            variant: 'success',
          },
        ]),
      )
      .catch(error => {
        if (isUnauthenticated(error.code)) {
          setShowAuthDialog(true);
        } else {
          setNotifications([{ message: error.message, variant: 'danger' }]);
        }
      });

  const cleanUp = useCallback(() => {
    setOpenDeletePR(false);
    setSelectedClusters([]);
  }, [setOpenDeletePR, setSelectedClusters]);

  useEffect(() => {
    if (
      notifications.length > 0 &&
      notifications[notifications.length - 1].variant !== 'danger' &&
      notifications[notifications.length - 1].message !==
        'Authentication completed successfully. Please proceed with creating the PR.'
    ) {
      cleanUp();
    }
  }, [notifications, setOpenDeletePR, setSelectedClusters, cleanUp]);

  return (
    <Dialog open maxWidth="md" fullWidth onClose={cleanUp}>
      <div id="delete-popup" className={classes.dialog}>
        <DialogTitle disableTypography>
          <Typography variant="h5">Create PR to remove clusters</Typography>
          <CloseIconButton onClick={cleanUp} />
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
          <GithubDeviceAuthModal
            onClose={() => setShowAuthDialog(false)}
            onSuccess={() => {
              setShowAuthDialog(false);
              setNotifications([
                {
                  message:
                    'Authentication completed successfully. Please proceed with creating the PR.',
                  variant: 'success',
                },
              ]);
            }}
            open={showAuthDialog}
            repoName="config"
          />
        </DialogContent>
      </div>
    </Dialog>
  );
};
