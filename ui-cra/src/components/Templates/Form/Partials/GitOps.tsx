import React, { FC, useCallback, Dispatch, ChangeEvent } from 'react';
import styled from 'styled-components';
import GitAuth from '../../../GitAuth';
import { Input } from '../../../../utils/form';
import { GitopsFormData } from '../utils';

const GitOpsWrapper = styled.div`
  padding-bottom: ${({ theme }) => theme.spacing.xl};
  .form-section {
    width: 50%;
  }
`;

const GitOps: FC<{
  formData: GitopsFormData;
  setFormData: Dispatch<React.SetStateAction<GitopsFormData>>;
  showAuthDialog: boolean;
  setShowAuthDialog: Dispatch<React.SetStateAction<boolean>>;
  setEnableCreatePR: Dispatch<React.SetStateAction<boolean>>;
  formError?: string;
  enableGitRepoSelection?: boolean;
}> = ({
  formData,
  setFormData,
  showAuthDialog,
  setShowAuthDialog,
  setEnableCreatePR,
  formError,
  enableGitRepoSelection,
}) => {
  const {
    branchName,
    pullRequestTitle,
    commitMessage,
    pullRequestDescription,
  } = formData;

  const handleChangeBranchName = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setFormData(prevState => ({
        ...prevState,
        branchName: event.target.value,
      })),
    [setFormData],
  );

  const handleChangePullRequestTitle = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setFormData(prevState => ({
        ...prevState,
        pullRequestTitle: event.target.value,
      })),
    [setFormData],
  );

  const handleChangeCommitMessage = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setFormData(prevState => ({
        ...prevState,
        commitMessage: event.target.value,
      })),
    [setFormData],
  );

  const handleChangePRDescription = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setFormData(prevState => ({
        ...prevState,
        pullRequestDescription: event.target.value,
      })),
    [setFormData],
  );

  return (
    <GitOpsWrapper>
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
      />
    </GitOpsWrapper>
  );
};

export default GitOps;
