import { Button } from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import { GitAuth } from '../../contexts/GitAuth';
import { Routes } from '../../utils/nav';
import { navigate } from './utils';

type Props = {
  className?: string;
  onClick: () => void;
};

function BitBucketAuthButton({ onClick, ...props }: Props) {
  const { gitAuthClient } = React.useContext(GitAuth);

  const handleClick = (e: any) => {
    gitAuthClient
      .GetBitbucketServerAuthURL({ redirectUri: redirectURI() })
      .then(res => {
        navigate(res?.url || '');
      });
    onClick();
  };
  return (
    <Button onClick={handleClick} {...props}>
      Authenticate with Bitbucket Server
    </Button>
  );
}

export default styled(BitBucketAuthButton).attrs({
  className: BitBucketAuthButton.name,
})``;

export function redirectURI(): string {
  return `${window.location.origin}${Routes.BitBucketOauthCallback}`;
}
