// eslint-disable-next-line
import { ButtonProps } from '@material-ui/core';
import {
  AppContext,
  Button,
  CallbackStateContext,
} from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import { gitlabOAuthRedirectURI } from '../../utils/request';

type Props = ButtonProps;

function GitlabAuthButton({ ...props }: Props) {
  const { callbackState } = React.useContext(CallbackStateContext);
  const { applicationsClient, navigate, storeCallbackState } =
    React.useContext(AppContext);

  const handleClick = (e: any) => {
    e.preventDefault();

    storeCallbackState(callbackState);

    applicationsClient
      .GetGitlabAuthURL({
        redirectUri: gitlabOAuthRedirectURI(),
      })
      .then((res: any) => {
        navigate.external(res.url);
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
