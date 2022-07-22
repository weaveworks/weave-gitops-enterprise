import React, { FC, useCallback, useState, Dispatch, ChangeEvent } from 'react';
import styled from 'styled-components';
import { Button, LoadingPage } from '@weaveworks/weave-gitops';
import GitAuth from './GitAuth';
import { Input, validateFormData } from '../../../../../utils/form';

const GitOpsWrapper = styled.form`
  padding-bottom: ${({ theme }) => theme.spacing.xl};
  .form-section {
    width: 50%;
  }
  .create-cta {
    display: flex;
    justify-content: end;
    padding: ${({ theme }) => theme.spacing.small};
    button {
      width: 200px;
    }
  }
`;

const GitOps: FC<{
  loading: boolean;
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  onSubmit: () => Promise<void>;
  showAuthDialog: boolean;
  setShowAuthDialog: Dispatch<React.SetStateAction<boolean>>;
}> = ({
  loading,
  formData,
  setFormData,
  onSubmit,
  showAuthDialog,
  setShowAuthDialog,
}) => {
  const [enableCreatePR, setEnableCreatePR] = useState<boolean>(false);

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

      {loading ? (
        <LoadingPage />
      ) : (
        <div className="create-cta">
          <Button
            onClick={event => validateFormData(event, onSubmit)}
            disabled={!enableCreatePR}
          >
            CREATE PULL REQUEST
          </Button>
        </div>
      )}
    </GitOpsWrapper>
  );
};

export default GitOps;
