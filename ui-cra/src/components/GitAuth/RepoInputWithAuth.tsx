import { MenuItem } from '@material-ui/core';
import {
  Button,
  Flex,
  Icon,
  IconType,
  useListSources,
} from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import { GitProvider } from '../../api/gitauth/gitauth.pb';
import { GitAuth, UseParseRepoUrl } from '../../contexts/GitAuth';
import { Select, SelectProps } from '../../utils/form';
import { getGitRepos } from '../Clusters';
import { getRepositoryUrl } from '../Templates/Form/utils';
import AuthButton from './AuthButton';

const GitAuthForm = styled(Flex)`
  #SELECT_GIT_REPO-group {
    width: 75%;
    padding-top: ${({ theme }) => theme.spacing.xs};
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
  const { data: res, error: err } = UseParseRepoUrl(value);
  const { data } = useListSources();
  const gitRepos = React.useMemo(
    () => getGitRepos(data?.result),
    [data?.result],
  );
  const { gitAuthClient } = React.useContext(GitAuth);

  const [valueForSelect, setValueForSelect] = React.useState<string>('');

  React.useEffect(() => {
    if (!value) {
      return;
    }
    setValueForSelect(value);

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [gitAuthClient, value]);

  React.useEffect(() => {
    if (!res) {
      return;
    }
    if (res.provider && onProviderChange) {
      onProviderChange(res.provider);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [res]);

  const renderProviderAuthButton =
    valueForSelect && !!res?.provider && !isAuthenticated;

  const handleSelectSource = (event: React.ChangeEvent<any>) => {
    const { value } = event.target;

    const gitRepo = gitRepos.find(repo => getRepositoryUrl(repo) === value);

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
          ?.map(gitRepo => ({'value': getRepositoryUrl(gitRepo), 'key': gitRepo.obj.spec.url }))
          .map((option, index: number) => (
            <MenuItem key={index} value={option.value}>
              {option.key}
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
        {renderProviderAuthButton && (
          <AuthButton provider={res?.provider} onClick={onAuthClick} />
        )}
      </div>
    </GitAuthForm>
  );
}
