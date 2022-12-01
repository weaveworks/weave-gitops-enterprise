import { ButtonProps } from '@material-ui/core';
import { Button } from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import { GitProvider } from '../../api/applications/applications.pb';
import { GitAuth } from '../../contexts/GitAuth';

type Props = ButtonProps;

function GithubAuthButton(props: Props) {
  return <Button {...props}>Authenticate with {GitProvider.GitHub}</Button>;
}

export default styled(GithubAuthButton).attrs({
  className: GithubAuthButton.name,
})`
  &.MuiButton-contained {
    background-color: black;
    color: white;
  }
`;

export function GlobalGithubAuthButton({ onSuccess }: any) {
  const {
    dialogState: { success },
    setDialogState,
  } = React.useContext(GitAuth);

  React.useEffect(() => {
    if (success && onSuccess) {
      onSuccess();
    }
  }, [success, onSuccess]);

  return <GithubAuthButton onClick={() => setDialogState(true, '')} />;
}
