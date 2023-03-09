import styled from 'styled-components';
import { GitProvider } from '../../api/gitauth/gitauth.pb';
import AzureServerAuthButton from './AzureServerAuthButton';
import BitBucketServerAuthButton from './BitBucketServerAuthButton';
import GithubAuthButton from './GithubAuthButton';
import GitlabAuthButton from './GitlabAuthButton';

type Props = {
  className?: string;
  provider?: GitProvider;
  onClick: (provider: GitProvider) => void;
};

function AuthButton({ className, provider, onClick, ...rest }: Props) {
  switch (provider) {
    case GitProvider.GitHub:
      return (
        <GithubAuthButton
          {...rest}
          onClick={() => onClick(GitProvider.GitHub)}
        />
      );

    case GitProvider.GitLab:
      return (
        <GitlabAuthButton
          {...rest}
          onClick={() => onClick(GitProvider.GitLab)}
        />
      );

    case GitProvider.BitBucketServer:
      return (
        <BitBucketServerAuthButton
          {...rest}
          onClick={() => onClick(GitProvider.BitBucketServer)}
        />
      );
    case GitProvider.AzureServer:
      return (
        <AzureServerAuthButton
          {...rest}
          onClick={() => onClick(GitProvider.AzureServer)}
        />
      );
    default:
      break;
  }
  return <div className={className}>Unknown Git Provider</div>;
}

export default styled(AuthButton).attrs({ className: AuthButton.name })``;
