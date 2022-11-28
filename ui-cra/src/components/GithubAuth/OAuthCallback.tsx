import { CircularProgress } from '@material-ui/core';
import {
  AppContext,
  AuthorizeGitlabResponse,
  Flex,
  useRequestState,
  useLinkResolver,
} from '@weaveworks/weave-gitops';
import * as React from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { GitProvider } from '../../contexts/GithubAuth/utils';
import { gitlabOAuthRedirectURI } from '../../utils/formatters';

type Props = {
  className?: string;
  code: string;
  provider: GitProvider;
};

function OAuthCallback({ className, code, provider }: Props) {
  const history = useHistory();
  const { applicationsClient, storeProviderToken, getCallbackState } =
    React.useContext(AppContext);
  const [res, loading, error, req] = useRequestState<AuthorizeGitlabResponse>();
  const linkResolver = useLinkResolver();

  React.useEffect(() => {
    if (provider === GitProvider.GitLab) {
      const redirectUri = gitlabOAuthRedirectURI();

      req(
        applicationsClient.AuthorizeGitlab({
          redirectUri,
          code,
        }),
      );
    }
  }, [code, applicationsClient, provider, req]);

  React.useEffect(() => {
    if (!res) {
      return;
    }

    storeProviderToken(GitProvider.GitLab, res.token || '');

    const state = getCallbackState();

    console.log(state.page);

    if (state?.page) {
      history.push(linkResolver(state.page));
      return;
    }
  }, [res]);

  return (
    // <Page className={className}>
    <Flex wide align center>
      {loading && <CircularProgress />}
      {/* {error && (
        <Alert
          title="Error completing OAuth 2.0 flow"
          severity="error"
          message={error.message}
        />
      )} */}
    </Flex>
    // </Page>
  );
}

export default styled(OAuthCallback).attrs({ className: OAuthCallback.name })``;
