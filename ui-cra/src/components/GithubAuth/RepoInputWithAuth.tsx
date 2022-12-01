import {
  Button,
  Flex,
  Icon,
  IconType,
  Input,
  InputProps,
  useRequestState,
  useDebounce,
  ParseRepoURLResponse,
} from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import GithubAuthButton from './GithubAuthButton';
import GitlabAuthButton from './GitlabAuthButton';
import { GitAuth } from '../../contexts/GitAuth';
import { GitProvider } from '../../api/applications/applications.pb';

type Props = InputProps & {
  onAuthClick: (provider: GitProvider) => void;
  onProviderChange?: (provider: GitProvider) => void;
  isAuthenticated?: boolean;
  disabled?: boolean;
};

function RepoInputWithAuth({
  onAuthClick,
  onProviderChange,
  isAuthenticated,
  disabled,
  ...props
}: Props) {
  const [res, , err, req] = useRequestState<ParseRepoURLResponse>();
  const debouncedURL = useDebounce<string>(props.value as string, 500);
  const { applicationsClient } = React.useContext(GitAuth);

  React.useEffect(() => {
    if (!debouncedURL) {
      return;
    }
    req(applicationsClient.ParseRepoURL({ url: debouncedURL }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    applicationsClient,
    debouncedURL,
    // req
    // is needed as dependency - infinite loop
  ]);

  React.useEffect(() => {
    if (!res) {
      return;
    }

    if (res.provider && onProviderChange) {
      onProviderChange(res.provider);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    res,
    // onProviderChange
  ]);

  const AuthButton =
    res?.provider === GitProvider.GitHub ? (
      <GithubAuthButton
        onClick={() => {
          onAuthClick(GitProvider.GitHub);
        }}
      />
    ) : (
      <GitlabAuthButton onClick={() => onAuthClick(GitProvider.GitLab)} />
    );

  const renderProviderAuthButton =
    props.value && !!res?.provider && !isAuthenticated;

  return (
    <Flex className={props.className} align start>
      <Input
        {...props}
        error={props.value && !!err?.message ? true : false}
        helperText={!props.value || !err ? props.helperText : err?.message}
        disabled={disabled}
      />
      <div className="auth-message">
        {isAuthenticated && (
          <Flex align>
            <Icon
              size="medium"
              color="successOriginal"
              type={IconType.CheckMark}
            />{' '}
            {res?.provider} credentials detected
          </Flex>
        )}
        {!isAuthenticated && !res && (
          <Button disabled>Authenticate with your Git Provider</Button>
        )}

        {renderProviderAuthButton ? AuthButton : null}
      </div>
    </Flex>
  );
}

export default styled(RepoInputWithAuth).attrs({
  className: RepoInputWithAuth.name,
})`
  .auth-message {
    margin-left: 8px;

    ${Icon} {
      margin-right: 4px;
    }
  }
`;
