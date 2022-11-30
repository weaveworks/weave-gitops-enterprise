import { ButtonProps } from '@material-ui/core';
import * as React from 'react';
import styled from 'styled-components';
import { Button, AppContext } from '@weaveworks/weave-gitops';
import { gitlabOAuthRedirectURI } from '../../utils/formatters';
import { CallbackStateContext } from '../../contexts/GithubAuth/CallbackStateContext';
import { storeCallbackState } from './utils';

type Props = ButtonProps;

function GitlabAuthButton({ ...props }: Props) {
  const { callbackState } = React.useContext(CallbackStateContext);
  const { applicationsClient, navigate } = React.useContext(AppContext);

  const handleClick = (e: any) => {
    e.preventDefault();

    storeCallbackState(callbackState);

    applicationsClient
      .GetGitlabAuthURL({
        redirectUri: gitlabOAuthRedirectURI(),
      })
      .then(res => {
        navigate.external(res.url || '');
      });
  };

  return (
    <Button {...props} onClick={handleClick}>
      Authenticate with GitLab
    </Button>
  );
}

export default styled(GitlabAuthButton).attrs({
  className: GitlabAuthButton.name,
})``;
