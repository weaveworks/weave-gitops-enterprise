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
import { CloseIconButton } from '../../assets/img/close-icon-button';
import useClusters from '../../contexts/Clusters';
import useNotifications from '../../contexts/Notifications';
import useVersions from '../../contexts/Versions';
import { faTrashAlt } from '@fortawesome/free-solid-svg-icons';
import { OnClickAction } from '../Action';
import { Input } from '../../utils/form';
import { Loader } from '../Loader';
import { clearCallbackState, getProviderToken } from '@weaveworks/weave-gitops';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import { isUnauthenticated, removeToken } from '../../utils/request';
import GitAuth from './Create/Form/Partials/GitAuth';

interface Props {
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
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
  formData,
  setFormData,
  selectedCapiClusters,
  setOpenDeletePR,
}) => {
  const classes = useStyles();
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const [enableCreatePR, setEnableCreatePR] = useState<boolean>(false);
  const { repositoryURL } = useVersions();

  const { deleteCreatedClusters, creatingPR, setSelectedClusters } =
    useClusters();
  const { notifications, setNotifications } = useNotifications();

  const handleChangeBranchName = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setFormData((prevState: any) => ({
        ...prevState,
        branchName: event.target.value,
      })),
    [setFormData],
  );

  const handleChangePullRequestTitle = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setFormData((prevState: any) => ({
        ...prevState,
        pullRequestTitle: event.target.value,
      })),
    [setFormData],
  );

  const handleChangeCommitMessage = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setFormData((prevState: any) => ({
        ...prevState,
        commitMessage: event.target.value,
      })),
    [setFormData],
  );

  const handleChangePRDescription = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setFormData((prevState: any) => ({
        ...prevState,
        pullRequestDescription: event.target.value,
      })),
    [setFormData],
  );

  const handleClickRemove = () =>
    deleteCreatedClusters(
      {
        clusterNames: selectedCapiClusters,
        headBranch: formData.branchName,
        title: formData.pullRequestTitle,
        commitMessage: formData.commitMessage,
        description: formData.pullRequestDescription,
        repositoryUrl: repositoryURL,
      },
      getProviderToken(formData.provider as GitProvider),
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
        setNotifications([{ message: error.message, variant: 'danger' }]);
        if (isUnauthenticated(error.code)) {
          removeToken(formData.provider);
        }
      });

  const cleanUp = useCallback(() => {
    setOpenDeletePR(false);
    setShowAuthDialog(false);
    setSelectedClusters([]);
    clearCallbackState();
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
  }, [
    notifications,
    setOpenDeletePR,
    setSelectedClusters,
    cleanUp,
    repositoryURL,
  ]);

  useEffect(() => {
    setFormData((prevState: FormData) => ({
      ...prevState,
      url: repositoryURL,
    }));
  }, [repositoryURL, setFormData]);

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
                placeholder={formData.branchName}
                onChange={handleChangeBranchName}
              />
              <Input
                label="Pull request title"
                placeholder={formData.pullRequestTitle}
                onChange={handleChangePullRequestTitle}
              />
              <Input
                label="Commit message"
                placeholder={formData.commitMessage}
                onChange={handleChangeCommitMessage}
              />
              <Input
                label="Pull request description"
                placeholder={formData.pullRequestDescription}
                onChange={handleChangePRDescription}
                multiline
                rows={4}
              />
              <GitAuth
                formData={formData}
                setFormData={setFormData}
                setEnableCreatePR={setEnableCreatePR}
                showAuthDialog={showAuthDialog}
                setShowAuthDialog={setShowAuthDialog}
              />
              <OnClickAction
                id="delete-cluster"
                icon={faTrashAlt}
                onClick={handleClickRemove}
                text="Remove clusters from the MCCP"
                className="danger"
                disabled={!enableCreatePR}
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
