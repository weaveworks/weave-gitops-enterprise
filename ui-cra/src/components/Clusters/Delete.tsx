import React, { ChangeEvent, FC, useCallback, useState, Dispatch } from 'react';
import {
  Dialog,
  DialogContent,
  DialogTitle,
  Typography,
} from '@material-ui/core';
import { CloseIconButton } from '../../assets/img/close-icon-button';
import useClusters from '../../contexts/Clusters';
import useNotifications from '../../contexts/Notifications';
import { Input } from '../../utils/form';
import { Loader } from '../Loader';
import {
  Button,
  clearCallbackState,
  getProviderToken,
  Icon,
  IconType,
} from '@weaveworks/weave-gitops';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import { isUnauthenticated, removeToken } from '../../utils/request';
import GitAuth from './Create/Form/Partials/GitAuth';
import { PRdefaults } from '.';

interface Props {
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  selectedCapiClusters: string[];
  setOpenDeletePR: Dispatch<React.SetStateAction<boolean>>;
}

export const DeleteClusterDialog: FC<Props> = ({
  formData,
  setFormData,
  selectedCapiClusters,
  setOpenDeletePR,
}) => {
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const [enableCreatePR, setEnableCreatePR] = useState<boolean>(false);

  const { deleteCreatedClusters, loading, setSelectedClusters } = useClusters();
  const { setNotifications } = useNotifications();

  // check the status of the clusters and exclude if they already have delete PR
  console.log(selectedCapiClusters);

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
        repositoryUrl: formData.repositoryURL,
      },
      getProviderToken(formData.provider as GitProvider),
    )
      .then(() => {
        cleanUp();
        setNotifications([
          {
            message: `PR created successfully`,
            variant: 'success',
          },
        ]);
      })
      .catch(error => {
        setNotifications([{ message: error.message, variant: 'danger' }]);
        if (isUnauthenticated(error.code)) {
          removeToken(formData.provider);
        }
      });

  const cleanUp = useCallback(() => {
    clearCallbackState();
    setShowAuthDialog(false);
    setSelectedClusters([]);
    setFormData(PRdefaults);
    setOpenDeletePR(false);
  }, [setSelectedClusters, setFormData, setOpenDeletePR]);

  return (
    <Dialog open maxWidth="md" fullWidth onClose={cleanUp}>
      <div id="delete-popup">
        <DialogTitle disableTypography>
          <Typography variant="h5">Create PR to remove clusters</Typography>
          <CloseIconButton onClick={cleanUp} />
        </DialogTitle>
        <DialogContent>
          {!loading ? (
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
              <Button
                id="delete-cluster"
                color="secondary"
                startIcon={<Icon type={IconType.DeleteIcon} size="base" />}
                onClick={handleClickRemove}
                disabled={!enableCreatePR}
              >
                REMOVE CLUSTERS FROM THE MCCP
              </Button>
            </>
          ) : (
            <Loader />
          )}
        </DialogContent>
      </div>
    </Dialog>
  );
};
