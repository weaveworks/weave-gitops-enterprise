import { ButtonProps } from '@material-ui/core';
import * as React from 'react';
import styled from 'styled-components';
import { Button } from '@weaveworks/weave-gitops';
import { gitlabOAuthRedirectURI } from '../../utils/formatters';
import {
  CallbackStateContext,
  CallbackStateContextType,
} from '../../contexts/GitAuth/CallbackStateContext';
import { navigate, storeCallbackState } from './utils';
import { GitAuth } from '../../contexts/GitAuth';

type Props = ButtonProps;

function GitlabAuthButton({ ...props }: Props) {
  const { callbackState } = React.useContext(
    CallbackStateContext as React.Context<CallbackStateContextType>,
  );
  const { applicationsClient } = React.useContext(GitAuth);

  const handleClick = (e: any) => {
    e.preventDefault();

    storeCallbackState(callbackState);

    applicationsClient
      .GetGitlabAuthURL({
        redirectUri: gitlabOAuthRedirectURI(),
      })
      .then(res => {
        navigate(res.url || '');
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
