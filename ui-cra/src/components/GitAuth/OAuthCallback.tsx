import { CircularProgress } from '@material-ui/core';
import { Alert, AlertTitle } from '@material-ui/lab';
import {
  Flex,
  Page,
  useLinkResolver,
  useRequestState,
} from '@weaveworks/weave-gitops';
import qs from 'query-string';
import * as React from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import {
  AuthorizeGitlabResponse,
  GitProvider,
} from '../../api/gitauth/gitauth.pb';
import { GitAuth } from '../../contexts/GitAuth';
import useNotifications from '../../contexts/Notifications';
import {
  azureDevOpsOAuthRedirectURI,
  bitbucketServerOAuthRedirectURI,
  gitlabOAuthRedirectURI,
} from '../../utils/formatters';
import { getCallbackState, storeProviderToken } from './utils';

type Props = {
  code: string;
  state: string;
  provider: GitProvider;
  error?: string | null;
  errorDescription?: string | null;
};

const ErrorMessage = ({ title, message }: any) => {
  return (
    <Alert severity="error">
      <AlertTitle>OAuth Error: {title} </AlertTitle>
      {message}
    </Alert>
  );
};

function OAuthCallback({
  code,
  state,
  provider,
  error: paramsError,
  errorDescription,
}: Props) {
  const history = useHistory();
  const [res, loading, error, req] = useRequestState<AuthorizeGitlabResponse>();
  const linkResolver = useLinkResolver();
  const { setNotifications } = useNotifications();
  const { gitAuthClient } = React.useContext(GitAuth);
  const params = qs.parse(history.location.search);

  React.useEffect(() => {
    if (provider === GitProvider.GitLab) {
      const redirectUri = gitlabOAuthRedirectURI();

      req(
        gitAuthClient.AuthorizeGitlab({
          redirectUri,
          code,
        }),
      );
    }

    if (provider === GitProvider.BitBucketServer) {
      req(
        gitAuthClient.AuthorizeBitbucketServer({
          code,
          state,
          redirectUri: bitbucketServerOAuthRedirectURI(),
        }),
      );
    }

    if (provider === GitProvider.AzureDevOps) {
      req(
        gitAuthClient.AuthorizeAzureDevOps({
          code,
          state,
          redirectUri: azureDevOpsOAuthRedirectURI(),
        }),
      );
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [code, gitAuthClient, provider]);

  React.useEffect(() => {
    if (!res) {
      return;
    }

    storeProviderToken(provider, res.token || '');

    const state = getCallbackState();

    if (state?.page || !params.error) {
      history.push(linkResolver(state?.page || ''));
      return;
    }
  }, [res, history, linkResolver, params.error, provider]);

  React.useEffect(() => {
    if (error) {
      setNotifications([{ message: { text: '' }, severity: 'error' }]);
    }
  }, [error, setNotifications]);

  return (
    <Page loading={loading} path={[{ label: 'OAuth Callback', url: '' }]}>
      <Flex wide align center>
        {loading && <CircularProgress />}
        {/* Two possible error sources: OAuth misconfiguration, 
            or a problem with the code exchange. Handling both here.
          */}
        {error && (
          <Alert severity="error">
            <AlertTitle>Request Error {error.name} </AlertTitle>
            {error.message}
          </Alert>
        )}
        {paramsError && (
          <ErrorMessage title={paramsError} message={errorDescription} />
        )}
      </Flex>
    </Page>
  );
}

export default styled(OAuthCallback).attrs({ className: OAuthCallback.name })``;
