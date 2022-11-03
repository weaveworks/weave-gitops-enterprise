import React, { ChangeEvent, FC, useCallback, useState, Dispatch } from 'react';
import {
  Dialog,
  DialogContent,
  DialogTitle,
  ThemeProvider,
  Typography,
} from '@material-ui/core';
import styled from 'styled-components';
import { CloseIconButton } from '../../assets/img/close-icon-button';
import useClusters from '../../hooks/clusters';
import { Input } from '../../utils/form';
import { Loader } from '../Loader';
import {
  Button,
  clearCallbackState,
  getProviderToken,
  Icon,
  IconType,
  Link,
} from '@weaveworks/weave-gitops';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import { isUnauthenticated, removeToken } from '../../utils/request';
import GitAuth from '../GithubAuth/GitAuth';
import { ClusterNamespacedName } from '../../cluster-services/cluster_services.pb';
import { NotificationData, PRDefaults } from '../../types/custom';
import { localEEMuiTheme } from '../../muiTheme';

const DeleteClusterWrapper = styled(Dialog)`
  #delete-popup {
    padding: 0 0 0 ${({ theme }) => theme.spacing.base};
  }
  h5 {
    padding-bottom: ${({ theme }) => theme.spacing.base};
  }
  .form-section {
    width: 100%;
  }
`;

interface Props {
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  selectedCapiClusters: ClusterNamespacedName[];
  onClose: (notification?: NotificationData) => void;
  prDefaults: PRDefaults;
}

export const DeleteClusterDialog: FC<Props> = ({
  formData,
  setFormData,
  selectedCapiClusters,
  onClose,
  prDefaults,
}) => {
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const [enableCreatePR, setEnableCreatePR] = useState<boolean>(false);
  const { deleteCreatedClusters, loading } = useClusters();

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
      .then(response =>
        cleanUp(
          {},
          {
            message: {
              component: (
                <Link href={response.webUrl} newTab>
                  PR created successfully.
                </Link>
              ),
            },
            severity: 'success',
          },
        ),
      )
      .catch(error => {
        cleanUp(
          {},
          {
            message: { text: error.message },
            severity: 'error',
          },
        );
        if (isUnauthenticated(error.code)) {
          removeToken(formData.provider);
        }
      });

  const cleanUp = useCallback(
    (event: any, notification?: NotificationData) => {
      clearCallbackState();
      setShowAuthDialog(false);
      setFormData(prDefaults);
      onClose(notification);
    },
    [onClose, setFormData, prDefaults],
  );

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <DeleteClusterWrapper
        open
        maxWidth="md"
        fullWidth
        onClose={event => cleanUp(event)}
      >
        <div id="delete-popup">
          <DialogTitle disableTypography>
            <Typography variant="h5">Create PR to remove clusters</Typography>
            <CloseIconButton onClick={event => cleanUp(event)} />
          </DialogTitle>
          <DialogContent>
            {!loading ? (
              <>
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
      </DeleteClusterWrapper>
    </ThemeProvider>
  );
};
