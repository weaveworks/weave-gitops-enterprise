import React, { FC, useCallback, Dispatch, ChangeEvent } from 'react';
import styled from 'styled-components';
import GitAuth from './GitAuth';
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
}> = ({
  formData,
  setFormData,
  showAuthDialog,
  setShowAuthDialog,
  setEnableCreatePR,
}) => {
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
        label="CREATE BRANCH"
        placeholder={formData.branchName}
        value={formData.branchName}
        onChange={handleChangeBranchName}
      />
      <Input
        className="form-section"
        required
        label="PULL REQUEST TITLE"
        placeholder={formData.pullRequestTitle}
        value={formData.pullRequestTitle}
        onChange={handleChangePullRequestTitle}
      />
      <Input
        className="form-section"
        required
        label="COMMIT MESSAGE"
        placeholder={formData.commitMessage}
        value={formData.commitMessage}
        onChange={handleChangeCommitMessage}
      />
      <Input
        className="form-section"
        required
        label="PULL REQUEST DESCRIPTION"
        placeholder={formData.pullRequestDescription}
        value={formData.pullRequestDescription}
        onChange={handleChangePRDescription}
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
