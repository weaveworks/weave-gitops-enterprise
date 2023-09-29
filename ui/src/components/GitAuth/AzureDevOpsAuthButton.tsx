// @ts-ignore
import { CallbackStateContextType } from '@weaveworks/weave-gitops/ui/contexts/CallbackStateContext';
import { Button } from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import { GitAuth } from '../../contexts/GitAuth';
import { CallbackStateContext } from '../../contexts/GitAuth/CallbackStateContext';
import { azureDevOpsOAuthRedirectURI } from '../../utils/formatters';
import { navigate, storeCallbackState } from './utils';

type Props = {
  className?: string;
  onClick: () => void;
};

function AzureDevOpsAuthButton({ onClick, ...props }: Props) {
  const { callbackState } = React.useContext<CallbackStateContextType>(
    CallbackStateContext as any,
  );
  const { gitAuthClient } = React.useContext(GitAuth);

  const handleClick = (e: any) => {
    storeCallbackState(callbackState);

    gitAuthClient
      .GetAzureDevOpsAuthURL({
        redirectUri: azureDevOpsOAuthRedirectURI(),
      })
      .then(res => {
        navigate(res?.url || '');
      });
    onClick();
  };
  return (
    <Button onClick={handleClick} {...props}>
      AUTHENTICATE WITH AZURE
    </Button>
  );
}

export default styled(AzureDevOpsAuthButton).attrs({
  className: AzureDevOpsAuthButton.name,
})``;
