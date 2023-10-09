import React, { ChangeEvent, Dispatch, FC, useCallback } from 'react';
import styled from 'styled-components';
import { Input } from '../../../../utils/form';
import GitAuth from '../../../GitAuth';
import { Button, Flex } from '@weaveworks/weave-gitops';

const GitOpsWrapper = styled.div`
  .form-section {
    width: 50%;
  }
`;

const GitOps: FC<{
  loading: boolean;
  isAuthenticated: boolean | undefined;
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  showAuthDialog: boolean;
  setShowAuthDialog: Dispatch<React.SetStateAction<boolean>>;
  formError?: string;
  enableGitRepoSelection?: boolean;
}> = ({
  loading,
  isAuthenticated,
  formData,
  setFormData,
  showAuthDialog,
  setShowAuthDialog,
  formError,
  enableGitRepoSelection,
  children,
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
      <h2>GitOps: Review and create</h2>
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
        showAuthDialog={showAuthDialog}
        setShowAuthDialog={setShowAuthDialog}
        enableGitRepoSelection={enableGitRepoSelection}
      />
      <Flex end className="gitops-cta">
        <Button
          loading={loading}
          type="submit"
          disabled={!isAuthenticated || loading}
        >
          CREATE PULL REQUEST
        </Button>
        {children}
      </Flex>
    </GitOpsWrapper>
  );
};

export default GitOps;
