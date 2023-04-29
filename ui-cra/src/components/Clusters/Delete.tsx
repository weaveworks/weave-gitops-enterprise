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
import useNotifications from '../../contexts/Notifications';
import { Input } from '../../utils/form';
import { Loader } from '../Loader';
import { Button, Icon, IconType, Link } from '@weaveworks/weave-gitops';
import { removeToken } from '../../utils/request';
import { ClusterNamespacedName } from '../../cluster-services/cluster_services.pb';
import { PRDefaults } from '../../types/custom';
import { localEEMuiTheme } from '../../muiTheme';
import GitAuth from '../GitAuth';
import { clearCallbackState, getProviderToken } from '../GitAuth/utils';
import { getRepositoryUrl } from '../Templates/Form/utils';
import {
  expiredTokenNotification,
  useIsAuthenticated,
} from '../../hooks/gitprovider';

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
  selectedCapiCluster: ClusterNamespacedName;
  onClose: () => void;
  prDefaults: PRDefaults;
}

export const DeleteClusterDialog: FC<Props> = ({
  formData,
  setFormData,
  selectedCapiCluster,
  onClose,
  prDefaults,
}) => {
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const { deleteCreatedClusters, loading } = useClusters();
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

  const token = getProviderToken(formData.provider);

  const { isAuthenticated, validateToken } = useIsAuthenticated(
    formData.provider,
    token,
  );

  const handleClickRemove = () =>
    validateToken()
      .then(() =>
        deleteCreatedClusters(
          {
            clusterNamespacedNames: [selectedCapiCluster],
            headBranch: formData.branchName,
            title: formData.pullRequestTitle,
            commitMessage: formData.commitMessage,
            description: formData.pullRequestDescription,
            repositoryUrl: getRepositoryUrl(formData.repo),
          },
          getProviderToken(formData.provider),
        )
          .then(response => {
            cleanUp();
            setNotifications([
              {
                message: {
                  component: (
                    <Link href={response.webUrl} newTab>
                      PR created successfully, please review and merge the pull
                      request to apply the changes to the cluster.
                    </Link>
                  ),
                },
                severity: 'success',
              },
            ]);
          })
          .catch(error =>
            setNotifications([
              { message: { text: error.message }, severity: 'error' },
            ]),
          ),
      )
      .catch(() => {
        removeToken(formData.provider);
        setNotifications([{ ...expiredTokenNotification, display: 'top' }]);
      });

  const cleanUp = useCallback(() => {
    clearCallbackState();
    setShowAuthDialog(false);
    setFormData(prDefaults);
    onClose();
  }, [onClose, setFormData, prDefaults]);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <DeleteClusterWrapper open maxWidth="md" fullWidth onClose={cleanUp}>
        <div id="delete-popup">
          <DialogTitle disableTypography>
            <Typography variant="h5">Create PR to remove clusters</Typography>
            <CloseIconButton onClick={cleanUp} />
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
                  showAuthDialog={showAuthDialog}
                  setShowAuthDialog={setShowAuthDialog}
                  enableGitRepoSelection={!formData?.repo?.createPRRepo}
                />
                <Button
                  id="delete-cluster"
                  color="secondary"
                  startIcon={<Icon type={IconType.DeleteIcon} size="base" />}
                  onClick={handleClickRemove}
                  disabled={!isAuthenticated}
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
