import React, { FC, useCallback, useState, Dispatch, ChangeEvent } from 'react';
import { Input } from '../../../../../utils/form';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { Button } from 'weaveworks-ui-components';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import { FormStep } from '../Step';
import GitAuth from './GitAuth';

const base = weaveTheme.spacing.base;

const useStyles = makeStyles(() =>
  createStyles({
    createCTA: {
      display: 'flex',
      justifyContent: 'center',
      paddingTop: base,
    },
  }),
);

const GitOps: FC<{
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  activeStep: string | undefined;
  setActiveStep: Dispatch<React.SetStateAction<string | undefined>>;
  clickedStep: string;
  setClickedStep: Dispatch<React.SetStateAction<string>>;
  onSubmit: (gitOps: {
    head_branch: string;
    title: string;
    description: string;
    commit_message: string;
  }) => Promise<void>;
  showAuthDialog: boolean;
  setShowAuthDialog: Dispatch<React.SetStateAction<boolean>>;
}> = ({
  formData,
  setFormData,
  activeStep,
  setActiveStep,
  clickedStep,
  setClickedStep,
  onSubmit,
  showAuthDialog,
  setShowAuthDialog,
}) => {
  const classes = useStyles();
  const [enableCreatePR, setEnableCreatePR] = useState<boolean>(false);

  const random = Math.random().toString(36).substring(7);

  const [branchName, setBranchName] = useState<string>(
    `create-clusters-branch-${random}`,
  );
  const [pullRequestTitle, setPullRequestTitle] = useState<string>(
    'Creates capi cluster',
  );
  const [commitMessage, setCommitMessage] = useState<string>(
    'Creates capi cluster',
  );
  const [pullRequestDescription, setPullRequestDescription] = useState<string>(
    'This PR creates a new cluster',
  );

  const handleChangeBranchName = useCallback(
    (event: ChangeEvent<HTMLInputElement>) => setBranchName(event.target.value),
    [],
  );

  const handleChangePullRequestTitle = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setPullRequestTitle(event.target.value),
    [],
  );

  const handleChangeCommitMessage = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setCommitMessage(event.target.value),
    [],
  );

  const handleChangePRDescription = useCallback(
    (event: ChangeEvent<HTMLInputElement>) =>
      setPullRequestDescription(event.target.value),
    [],
  );

  const handleGitOps = useCallback(
    () =>
      onSubmit({
        head_branch: branchName,
        title: pullRequestTitle,
        description: pullRequestDescription,
        commit_message: commitMessage,
      }),
    [
      branchName,
      pullRequestTitle,
      pullRequestDescription,
      commitMessage,
      onSubmit,
    ],
  );

  return (
    <FormStep
      title="GitOps"
      active={activeStep === 'GitOps'}
      clicked={clickedStep === 'GitOps'}
      setActiveStep={setActiveStep}
    >
      <Input
        label="Create branch"
        placeholder={branchName}
        onChange={handleChangeBranchName}
      />
      <Input
        label="Pull request title"
        placeholder={pullRequestTitle}
        onChange={handleChangePullRequestTitle}
      />
      <Input
        label="Commit message"
        placeholder={commitMessage}
        onChange={handleChangeCommitMessage}
      />
      <Input
        label="Pull request description"
        placeholder={pullRequestDescription}
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
      <div className={classes.createCTA} onClick={handleGitOps}>
        <Button
          onClick={() => setClickedStep('GitOps')}
          disabled={!enableCreatePR}
        >
          Create Pull Request
        </Button>
      </div>
    </FormStep>
  );
};

export default GitOps;
