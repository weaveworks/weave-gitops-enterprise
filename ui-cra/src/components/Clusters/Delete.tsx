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
import {
  CallbackStateContextProvider,
  getCallbackState,
  getProviderToken,
} from '@weaveworks/weave-gitops';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import { isUnauthenticated } from '../../utils/request';
import GitAuth from './Create/Form/Partials/GitAuth';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';

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
  const [enableCreatePR, setEnableCreatePR] = useState<boolean>(false);
  const { repositoryURL } = useVersions();
  const authRedirectPage = `/clusters/delete`;

  interface FormData {
    url: string;
    provider: string;
    branchName: string;
    pullRequestTitle: string;
    commitMessage: string;
    pullRequestDescription: string;
  }

  let initialFormData = {
    url: repositoryURL,
    provider: '',
    branchName: `delete-clusters-branch-${random}`,
    pullRequestTitle: 'Deletes capi cluster(s)',
    commitMessage: 'Deletes capi cluster(s)',
    pullRequestDescription: `Delete clusters: ${selectedCapiClusters
      .map(c => c)
      .join(', ')}`,
  };

  const callbackState = getCallbackState();

  if (callbackState) {
    initialFormData = {
      ...initialFormData,
      ...callbackState.state,
    };
    console.log(callbackState);
    // setOpenDeletePR(true);
  }

  const [formData, setFormData] = useState<FormData>(initialFormData);

  const { deleteCreatedClusters, creatingPR, setSelectedClusters } =
    useClusters();
  const { notifications, setNotifications } = useNotifications();

  const handleChangeBranchName = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setFormData((prevState: any) => ({
        ...prevState,
        branchName: event.target.value,
      })),
    [],
  );

  const handleChangePullRequestTitle = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setFormData((prevState: any) => ({
        ...prevState,
        pullRequestTitle: event.target.value,
      })),
    [],
  );

  const handleChangeCommitMessage = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setFormData((prevState: any) => ({
        ...prevState,
        commitMessage: event.target.value,
      })),
    [],
  );

  const handleChangePRDescription = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setFormData((prevState: any) => ({
        ...prevState,
        pullRequestDescription: event.target.value,
      })),
    [],
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
        if (isUnauthenticated(error.code)) {
          setShowAuthDialog(true);
        } else {
          setNotifications([{ message: error.message, variant: 'danger' }]);
        }
      });

  const cleanUp = useCallback(() => {
    setOpenDeletePR(false);
    setShowAuthDialog(false);
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
  }, [repositoryURL]);

  return (
    <Dialog open maxWidth="md" fullWidth onClose={cleanUp}>
      <CallbackStateContextProvider
        callbackState={{
          page: authRedirectPage as PageRoute,
          state: formData,
        }}
      >
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
                  disabled={
                    selectedCapiClusters.length === 0 || !enableCreatePR
                  }
                />
              </>
            ) : (
              <Loader />
            )}
          </DialogContent>
        </div>
      </CallbackStateContextProvider>
    </Dialog>
  );
};
