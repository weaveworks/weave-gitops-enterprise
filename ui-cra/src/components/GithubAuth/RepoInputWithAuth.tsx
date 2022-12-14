import {
  Button,
  Flex,
  Icon,
  IconType,
  useRequestState,
  useDebounce,
  GitRepository,
  useListSources,
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
import { Select, SelectProps } from '../../utils/form';
import { MenuItem } from '@material-ui/core';
import { getGitRepos } from '../Clusters';

const GitAuthForm = styled(Flex)`
  div[class*='MuiFormControl-root'] {
    padding-bottom: 0;
  }
`;

type Props = SelectProps & {
  onAuthClick: (provider: GitProvider) => void;
  onProviderChange?: (provider: GitProvider) => void;
  isAuthenticated?: boolean;
  disabled?: boolean;
  formData: any;
  setFormData: React.Dispatch<React.SetStateAction<any>>;
};

function RepoInputWithAuth({
  onAuthClick,
  onProviderChange,
  isAuthenticated,
  disabled,
  formData,
  setFormData,
  ...props
}: Props) {
  const [res, , err, req] = useRequestState<ParseRepoURLResponse>();
  const { data } = useListSources();
  const gitRepos = React.useMemo(
    () => getGitRepos(data?.result),
    [data?.result],
  );
  const value = gitRepos.filter(
    gitRepo => gitRepo.clusterName === 'management',
  )[0]?.obj?.spec?.url;
  // const debouncedURL = useDebounce<string>(value, 500);
  const { gitAuthClient } = React.useContext(GitAuth);

  console.log(gitRepos);

  React.useEffect(() => {
    // if (!debouncedURL) {
    //   return;
    // }
    if (!value) {
      return;
    }
    // req(gitAuthClient.ParseRepoURL({ url: debouncedURL }));
    req(gitAuthClient.ParseRepoURL({ url: value }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    gitAuthClient,
    value,
    // debouncedURL
  ]);

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

  const renderProviderAuthButton = value && !!res?.provider && !isAuthenticated;

  const handleSelectSource = (event: React.ChangeEvent<any>) => {
    const { value } = event.target;
    const { obj } = JSON.parse(value);

    setFormData({
      ...formData,
      url: value,
    });
  };

  console.log(formData.url);

  return (
    <GitAuthForm className={props.className} align start>
      <Select
        error={gitRepos && !!err?.message ? true : false}
        description={!value || !err ? props.description : err?.message}
        name="repo-select"
        required={true}
        label="SELECT GIT REPO"
        value={JSON.stringify(formData.url)}
        onChange={handleSelectSource}
      >
        {gitRepos?.map((option, index: number) => (
          <MenuItem key={index} value={JSON.stringify(option)}>
            {option?.obj?.spec?.url}
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
    </GitAuthForm>
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
