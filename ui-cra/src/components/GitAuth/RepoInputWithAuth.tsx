import {
  Button,
  Flex,
  Icon,
  IconType,
  useRequestState,
  useListSources,
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
import { Select, SelectProps } from '../../utils/form';
import { MenuItem } from '@material-ui/core';
import { getGitRepos } from '../Clusters';

const getUrlFromRepo = (repo: GitRepository | null) => repo?.obj?.spec?.url;

const GitAuthForm = styled(Flex)`
  div[class*='MuiFormControl-root'] {
    padding-bottom: 0;
  }
  justify-content: flex-start;
  #SELECT_GIT_REPO-group {
    width: 70%;
  }
`;

type Props = SelectProps & {
  onAuthClick: (provider: GitProvider) => void;
  onProviderChange?: (provider: GitProvider) => void;
  isAuthenticated?: boolean;
  disabled?: boolean;
  formData: any;
  setFormData: React.Dispatch<React.SetStateAction<any>>;
  enableGitRepoSelection?: boolean;
  value: string;
};

export function RepoInputWithAuth({
  onAuthClick,
  onProviderChange,
  isAuthenticated,
  disabled,
  formData,
  setFormData,
  enableGitRepoSelection,
  value,
  ...props
}: Props) {
  const [res, , err, req] = useRequestState<ParseRepoURLResponse>();
  const { data } = useListSources();
  const gitRepos = React.useMemo(
    () => getGitRepos(data?.result),
    [data?.result],
  );
  const { gitAuthClient } = React.useContext(GitAuth);

  const getDefaultValue = React.useCallback(() => {
    const annoRepo = gitRepos.find(
      repo =>
        repo?.obj?.metadata?.annotations?.['weave.works/repo-role'] ===
        'default',
    );
    if (annoRepo) {
      return getUrlFromRepo(annoRepo);
    }
    const mainRepo = gitRepos.find(
      repo =>
        repo?.obj?.metadata?.name === 'flux-system' &&
        repo?.obj?.metadata?.namespace === 'flux-system',
    );
    if (mainRepo) {
      return getUrlFromRepo(mainRepo);
    }
  }, [gitRepos]);

  const [valueForSelect, setValueForSelect] = React.useState<string>('');

  React.useEffect(() => {
    const defaultValue = getDefaultValue();

    if (!value && !defaultValue) {
      return;
    }

    setValueForSelect(value || defaultValue);

    req(
      gitAuthClient.ParseRepoURL({
        url: value || defaultValue,
      }),
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [gitAuthClient, value, getDefaultValue]);

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
    valueForSelect && !!res?.provider && !isAuthenticated;

  const handleSelectSource = (event: React.ChangeEvent<any>) => {
    const { value } = event.target;

    const gitRepo = gitRepos.find(repo => getUrlFromRepo(repo) === value);

    setFormData((prevState: any) => {
      return {
        ...prevState,
        repo: gitRepo,
      };
    });
  };

  return (
    <GitAuthForm className={props.className} align start>
      <Select
        error={gitRepos && !!err?.message ? true : false}
        description={!formData.repo || !err ? props.description : err?.message}
        name="repo-select"
        required={true}
        label="SELECT_GIT_REPO"
        value={valueForSelect}
        onChange={handleSelectSource}
        disabled={!enableGitRepoSelection}
      >
        {gitRepos
          ?.map(gitRepo => gitRepo.obj.spec.url)
          .map((option, index: number) => (
            <MenuItem key={index} value={option}>
              {option}
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
