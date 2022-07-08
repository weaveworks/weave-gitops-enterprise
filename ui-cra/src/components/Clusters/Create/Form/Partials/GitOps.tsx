import React, { FC, useCallback, useState, Dispatch, ChangeEvent } from 'react';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { Button } from '@weaveworks/weave-gitops';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import GitAuth from './GitAuth';
import FormControl from '@material-ui/core/FormControl';
import Input from '@material-ui/core/Input';

const base = weaveTheme.spacing.base;

const useStyles = makeStyles(() =>
  createStyles({
    createCTA: {
      display: 'flex',
      justifyContent: 'end',
      paddingTop: base,
    },
    formField: {
      width: '50%',
      paddingBottom: weaveTheme.spacing.small,
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
    <div style={{ paddingBottom: weaveTheme.spacing.xl }}>
      <h2>GitOps</h2>
      <FormControl className={classes.formField}>
        <span>CREATE BRANCH</span>
        <Input
          id="Create branch"
          placeholder={formData.branchName}
          value={formData.branchName}
          onChange={handleChangeBranchName}
        />
      </FormControl>
      <FormControl className={classes.formField}>
        <span>PULL REQUEST TITLE</span>
        <Input
          id="Pull request title"
          placeholder={formData.pullRequestTitle}
          value={formData.pullRequestTitle}
          onChange={handleChangePullRequestTitle}
        />
      </FormControl>
      <FormControl className={classes.formField}>
        <span>COMMIT MESSAGE</span>
        <Input
          id="Commit message"
          placeholder={formData.commitMessage}
          value={formData.commitMessage}
          onChange={handleChangeCommitMessage}
        />
      </FormControl>
      <FormControl className={classes.formField}>
        <span>PULL REQUEST DESCRIPTION</span>
        <Input
          id="Commit message"
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
      <div className={classes.createCTA} onClick={handleGitOps}>
        <Button disabled={!enableCreatePR}>CREATE PULL REQUEST</Button>
      </div>
    </div>
  );
};

export default GitOps;
