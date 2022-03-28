import { CircularProgress } from '@material-ui/core';
import * as React from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import { gitlabOAuthRedirectURI } from '../../utils/request';
import {
  Alert,
  AppContext,
  AuthorizeGitlabResponse,
  Flex,
  useRequestState,
} from '@weaveworks/weave-gitops';

type Props = {
  className?: string;
  code: string;
  provider: GitProvider;
};

function OAuthCallback({ code, provider }: Props) {
  const history = useHistory();
  const {
    applicationsClient,
    storeProviderToken,
    getCallbackState,
    linkResolver,
  } = React.useContext(AppContext);
  const [res, loading, error, req] = useRequestState<AuthorizeGitlabResponse>();

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
  }, [code]);

  React.useEffect(() => {
    if (!res) {
      return;
    }

    //@ts-ignore
    storeProviderToken(GitProvider.GitLab, res.token);

    const state = getCallbackState();

    if (state?.page) {
      history.push(linkResolver(state.page));
      return;
    }
  }, [res]);

  return (
    <Flex wide align center>
      {loading && <CircularProgress />}
      {error && (
        <Alert
          title="Error completing OAuth 2.0 flow"
          severity="error"
          message={error.message}
        />
      )}
    </Flex>
  );
}

export default styled(OAuthCallback).attrs({ className: OAuthCallback.name })``;
