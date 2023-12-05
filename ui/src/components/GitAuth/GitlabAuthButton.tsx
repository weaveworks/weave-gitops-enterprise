import { ButtonProps } from '@material-ui/core';
import { Button } from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import { useEnterpriseClient } from '../../contexts/API';
import {
  CallbackStateContext,
  CallbackStateContextType,
} from '../../contexts/GitAuth/CallbackStateContext';
import { gitlabOAuthRedirectURI } from '../../utils/formatters';
import { navigate, storeCallbackState } from './utils';

type Props = ButtonProps;

function GitlabAuthButton({ ...props }: Props) {
  const { callbackState } = React.useContext(
    CallbackStateContext as React.Context<CallbackStateContextType>,
  );
  const { gitAuth } = useEnterpriseClient();

  const handleClick = (e: any) => {
    e.preventDefault();

    storeCallbackState(callbackState);

    gitAuth
      .GetGitlabAuthURL({
        redirectUri: gitlabOAuthRedirectURI(),
      })
      .then(res => {
        navigate(res.url || '');
      });
  };

  return (
    <Button {...props} onClick={handleClick}>
      AUTHENTICATE WITH GITLAB
    </Button>
  );
}

export default styled(GitlabAuthButton).attrs({
  className: GitlabAuthButton.name,
})``;
