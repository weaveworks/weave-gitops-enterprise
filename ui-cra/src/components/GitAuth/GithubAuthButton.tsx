// eslint-disable-next-line
import { ButtonProps } from '@material-ui/core';
import { Button } from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import { GithubAuthContext } from '../../contexts/GitAuth/GithubAuthContext';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';

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

//@ts-ignore
export function GlobalGithubAuthButton({ onSuccess }) {
  const {
    dialogState: { success },
    setDialogState,
  } = React.useContext(GithubAuthContext);

  React.useEffect(() => {
    if (success && onSuccess) {
      onSuccess();
    }
  }, [success]);

  return <GithubAuthButton onClick={() => setDialogState(true, '')} />;
}
