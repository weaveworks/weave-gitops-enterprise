// @ts-ignore
import { CallbackStateContextType } from '@weaveworks/weave-gitops/ui/contexts/CallbackStateContext';
import { GitAuth } from '../../contexts/GitAuth';
import { CallbackStateContext } from '../../contexts/GitAuth/CallbackStateContext';
import { bitbucketServerOAuthRedirectURI } from '../../utils/formatters';
import { navigate, storeCallbackState } from './utils';
import { Button } from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';

type Props = {
  className?: string;
  onClick: () => void;
};

function BitBucketAuthButton({ onClick, ...props }: Props) {
  const { callbackState } = React.useContext<CallbackStateContextType>(
    CallbackStateContext as any,
  );
  const { gitAuthClient } = React.useContext(GitAuth);

  const handleClick = (e: any) => {
    storeCallbackState(callbackState);

    gitAuthClient
      .GetBitbucketServerAuthURL({
        redirectUri: bitbucketServerOAuthRedirectURI(),
      })
      .then(res => {
        navigate(res?.url || '');
      });
    onClick();
  };
  return (
    <Button onClick={handleClick} {...props}>
      AUTHENTICATE WITH BITBUCKET
    </Button>
  );
}

export default styled(BitBucketAuthButton).attrs({
  className: BitBucketAuthButton.name,
})``;
