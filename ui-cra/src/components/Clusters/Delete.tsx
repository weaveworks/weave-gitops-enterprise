import React, { ChangeEvent, FC, useCallback, useState, Dispatch } from 'react';
import {
  Dialog,
  DialogContent,
  DialogTitle,
  Typography,
} from '@material-ui/core';
import styled from 'styled-components';
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
  theme,
} from '@weaveworks/weave-gitops';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import { isUnauthenticated, removeToken } from '../../utils/request';
import GitAuth from './Create/Form/Partials/GitAuth';
import { PRdefaults } from '.';
import { ClusterNamespacedName } from '../../cluster-services/cluster_services.pb';

const FormWrapper = styled.form`
  .form-section {
    width: 100%;
  }
`;

interface Props {
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  selectedCapiClusters: ClusterNamespacedName[];
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
        clusterNamespacedNames: selectedCapiClusters,
        headBranch: formData.branchName,
        title: formData.pullRequestTitle,
        commitMessage: formData.commitMessage,
        description: formData.pullRequestDescription,
        repositoryUrl: formData.repositoryURL,
      },
      getProviderToken(formData.provider as GitProvider),
    )
      .then(response => {
        cleanUp();
        setNotifications([
          {
            message: {
              component: (
                <a
                  style={{ color: theme.colors.primary }}
                  href={response.webUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  PR created successfully.
                </a>
              ),
            },
            variant: 'success',
          },
        ]);
      })
      .catch(error => {
        setNotifications([
          { message: { text: error.message }, variant: 'danger' },
        ]);
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
            <FormWrapper>
              <Input
                className="form-section"
                label="CREATE BRANCH"
                placeholder={formData.branchName}
                onChange={handleChangeBranchName}
              />
              <Input
                className="form-section"
                label="PULL REQUEST TITLE"
                placeholder={formData.pullRequestTitle}
                onChange={handleChangePullRequestTitle}
              />
              <Input
                className="form-section"
                label="COMMIT MESSAGE"
                placeholder={formData.commitMessage}
                onChange={handleChangeCommitMessage}
              />
              <Input
                className="form-section"
                label="PULL REQUEST DESCRIPTION"
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
            </FormWrapper>
          ) : (
            <Loader />
          )}
        </DialogContent>
      </div>
    </Dialog>
  );
};
