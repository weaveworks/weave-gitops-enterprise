import React, { FC, useCallback, useState, Dispatch, ChangeEvent } from 'react';
import { Input } from '../../../../../utils/form';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { Button } from '@weaveworks/weave-gitops';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';
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
  const classes = useStyles();
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

  const handleGitOps = useCallback(() => onSubmit(), [onSubmit]);

  return (
    <>
      <Input
        label="Create branch"
        placeholder={formData.branchName}
        onChange={handleChangeBranchName}
      />
      <Input
        label="Pull request title"
        placeholder={formData.pullRequestTitle}
        onChange={handleChangePullRequestTitle}
      />
      <Input
        label="Commit message"
        placeholder={formData.commitMessage}
        onChange={handleChangeCommitMessage}
      />
      <Input
        label="Pull request description"
        placeholder={formData.pullRequestDescription}
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
        <Button disabled={!enableCreatePR}>CREATE PULL REQUEST</Button>
      </div>
    </>
  );
};

export default GitOps;
