import React, { FC, useCallback, Dispatch, ChangeEvent } from 'react';
import styled from 'styled-components';
import GitAuth from '../../../GithubAuth/GitAuth';
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
}> = ({
  formData,
  setFormData,
  showAuthDialog,
  setShowAuthDialog,
  setEnableCreatePR,
  formError,
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
        error={
          formError === 'branch_name' &&
          (branchName === undefined || branchName === '' || branchName === null)
        }
      />
      <Input
        className="form-section"
        required
        name="pull_request_title"
        label="PULL REQUEST TITLE"
        placeholder={pullRequestTitle}
        value={pullRequestTitle}
        onChange={handleChangePullRequestTitle}
        error={
          formError === 'pull_request_title' &&
          (pullRequestTitle === undefined || pullRequestTitle === '')
        }
      />
      <Input
        className="form-section"
        required
        name="commit_message"
        label="COMMIT MESSAGE"
        placeholder={commitMessage}
        value={commitMessage}
        onChange={handleChangeCommitMessage}
        error={
          formError === 'commit_message' &&
          (commitMessage === undefined || commitMessage === '')
        }
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
          formError === 'pull_request_description' &&
          (pullRequestDescription === undefined ||
            pullRequestDescription === '')
        }
      />
      <GitAuth
        formData={formData}
        setFormData={setFormData}
        setEnableCreatePR={setEnableCreatePR}
        showAuthDialog={showAuthDialog}
        setShowAuthDialog={setShowAuthDialog}
      />
    </GitOpsWrapper>
  );
};

export default GitOps;
