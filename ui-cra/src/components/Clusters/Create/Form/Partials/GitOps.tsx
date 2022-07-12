import React, { FC, useCallback, useState, Dispatch, ChangeEvent } from 'react';
import styled from 'styled-components';
import { Button } from '@weaveworks/weave-gitops';
import GitAuth from './GitAuth';
import FormControl from '@material-ui/core/FormControl';
import Input from '@material-ui/core/Input';
import { validateFormData } from '../../../../../utils/form';

const GitOpsWrapper = styled.form`
  padding-bottom: ${({ theme }) => theme.spacing.xl};
  .form-field {
    width: 50%;
    padding-bottom: ${({ theme }) => theme.spacing.small};
  }
  .create-cta {
    display: flex;
    justify-content: end;
    padding-top: ${({ theme }) => theme.spacing.base};
    button {
      width: 200px;
    }
  }
`;

const GitOps: FC<{
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  onSubmit: () => Promise<void>;
  showAuthDialog: boolean;
  setShowAuthDialog: Dispatch<React.SetStateAction<boolean>>;
}> = ({
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
      <FormControl className="form-field">
        <span>CREATE BRANCH</span>
        <Input
          id="Create branch"
          required
          placeholder={formData.branchName}
          value={formData.branchName}
          onChange={handleChangeBranchName}
        />
      </FormControl>
      <FormControl className="form-field">
        <span>PULL REQUEST TITLE</span>
        <Input
          id="Pull request title"
          required
          placeholder={formData.pullRequestTitle}
          value={formData.pullRequestTitle}
          onChange={handleChangePullRequestTitle}
        />
      </FormControl>
      <FormControl className="form-field">
        <span>COMMIT MESSAGE</span>
        <Input
          id="Commit message"
          required
          placeholder={formData.commitMessage}
          value={formData.commitMessage}
          onChange={handleChangeCommitMessage}
        />
      </FormControl>
      <FormControl className="form-field">
        <span>PULL REQUEST DESCRIPTION</span>
        <Input
          id="Commit message"
          required
          placeholder={formData.pullRequestDescription}
          value={formData.pullRequestDescription}
          onChange={handleChangePRDescription}
        />
      </FormControl>
      <GitAuth
        formData={formData}
        setFormData={setFormData}
        setEnableCreatePR={setEnableCreatePR}
        showAuthDialog={showAuthDialog}
        setShowAuthDialog={setShowAuthDialog}
      />
      <div className="create-cta">
        <Button
          onClick={event => validateFormData(event, onSubmit)}
          disabled={!enableCreatePR}
        >
          CREATE PULL REQUEST
        </Button>
      </div>
    </GitOpsWrapper>
  );
};

export default GitOps;
