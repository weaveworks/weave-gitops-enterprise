import React, { FC, useCallback, Dispatch, ChangeEvent } from 'react';
import styled from 'styled-components';
import GitAuth from '../../../GitAuth';
import { Input } from '../../../../utils/form';

const GitOpsWrapper = styled.div`
  padding-bottom: ${({ theme }) => theme.spacing.xl};
  .form-section {
    width: 50%;
  }
`;

const GitOps: FC<{
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  showAuthDialog: boolean;
  setShowAuthDialog: Dispatch<React.SetStateAction<boolean>>;
  setEnableCreatePR: Dispatch<React.SetStateAction<boolean>>;
  formError?: string;
  enableGitRepoSelection?: boolean;
  creatingPR?: boolean;
  setCreatingPR?: Dispatch<React.SetStateAction<boolean>>;
  setSendPR?: Dispatch<React.SetStateAction<boolean>>;
}> = ({
  formData,
  setFormData,
  showAuthDialog,
  setShowAuthDialog,
  setEnableCreatePR,
  formError,
  enableGitRepoSelection,
  creatingPR,
  setCreatingPR,
  setSendPR,
}) => {
  const {
    branchName,
    pullRequestTitle,
    commitMessage,
    pullRequestDescription,
  } = formData;

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

  return (
    <GitOpsWrapper className="gitops-wrapper">
      <h2>GitOps</h2>
      <Input
        className="form-section"
        required
        name="branch_name"
        label="CREATE BRANCH"
        placeholder={branchName}
        value={branchName}
        onChange={handleChangeBranchName}
        error={formError === 'branch_name' && !branchName}
      />
      <Input
        className="form-section"
        required
        name="pull_request_title"
        label="PULL REQUEST TITLE"
        placeholder={pullRequestTitle}
        value={pullRequestTitle}
        onChange={handleChangePullRequestTitle}
        error={formError === 'pull_request_title' && !pullRequestTitle}
      />
      <Input
        className="form-section"
        required
        name="commit_message"
        label="COMMIT MESSAGE"
        placeholder={commitMessage}
        value={commitMessage}
        onChange={handleChangeCommitMessage}
        error={formError === 'commit_message' && !commitMessage}
      />
      <Input
        className="form-section"
        required
        name="pull_request_description"
        label="PULL REQUEST DESCRIPTION"
        placeholder={pullRequestDescription}
        value={pullRequestDescription}
        onChange={handleChangePRDescription}
        error={
          formError === 'pull_request_description' && !pullRequestDescription
        }
      />
      <GitAuth
        formData={formData}
        setFormData={setFormData}
        setEnableCreatePR={setEnableCreatePR}
        showAuthDialog={showAuthDialog}
        setShowAuthDialog={setShowAuthDialog}
        enableGitRepoSelection={enableGitRepoSelection}
        creatingPR={creatingPR}
        setCreatingPR={setCreatingPR}
        setSendPR={setSendPR}
      />
    </GitOpsWrapper>
  );
};

export default GitOps;
