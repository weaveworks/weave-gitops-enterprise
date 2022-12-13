import {
  Button,
  Flex,
  Icon,
  IconType,
  InputProps,
  useRequestState,
  useDebounce,
  GitRepository,
} from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import GithubAuthButton from './GithubAuthButton';
import GitlabAuthButton from './GitlabAuthButton';
import { GitAuth } from '../../contexts/GitAuth';
import {
  GitProvider,
  ParseRepoURLResponse,
} from '../../api/gitauth/gitauth.pb';
import { Select } from '../../utils/form';
import MenuItem from '@material-ui/core/MenuItem';

type Props = {
  className?: string;
  onAuthClick: (provider: GitProvider) => void;
  onProviderChange?: (provider: GitProvider) => void;
  isAuthenticated?: boolean;
  disabled?: boolean;
  values: GitRepository[];
};

function RepoInputWithAuth({
  onAuthClick,
  onProviderChange,
  isAuthenticated,
  disabled,
  ...props
}: Props) {
  const [res, , err, req] = useRequestState<ParseRepoURLResponse>();
  const { values } = props;
  const debouncedURL = useDebounce<string>(props.values?.[0] as string, 500);
  const { gitAuthClient } = React.useContext(GitAuth);

  React.useEffect(() => {
    if (!debouncedURL) {
      return;
    }
    req(gitAuthClient.ParseRepoURL({ url: debouncedURL }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [gitAuthClient, debouncedURL]);

  React.useEffect(() => {
    if (!res) {
      return;
    }

    if (res.provider && onProviderChange) {
      onProviderChange(res.provider);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [res]);

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

  const handleSelectSource = (event: React.ChangeEvent<any>) => {
    const { value } = event.target;
    const { obj } = JSON.parse(value);

    setFormData({
      ...formData,
      url: gitRepo,
    });
  };

  return (
    <Flex className={props.className} align start>
      <Select
        {...props}
        error={props.value && !!err?.message ? true : false}
        // helperText={!props.value || !err ? props.helperText : err?.message}
        name="repo-select"
        required={true}
        label="SELECT GIT REPO"
        value={values[0] || ''}
        onChange={handleSelectSource}
        defaultValue={''}
        description="The name and type of source"
      >
        {props.values?.map((option, index: number) => (
          <MenuItem key={index} value={JSON.stringify(option)}>
            {option.name}
          </MenuItem>
        ))}
      </Select>
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
