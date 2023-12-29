import { CircularProgress } from '@material-ui/core';
import { Alert, AlertTitle } from '@material-ui/lab';
import { Flex, useLinkResolver } from '@weaveworks/weave-gitops';
import qs from 'query-string';
import * as React from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import {
  AuthorizeGitlabResponse,
  GitProvider,
} from '../../api/gitauth/gitauth.pb';
import { useEnterpriseClient } from '../../contexts/API';
import useNotifications from '../../contexts/Notifications';
import {
  azureDevOpsOAuthRedirectURI,
  bitbucketServerOAuthRedirectURI,
  gitlabOAuthRedirectURI,
} from '../../utils/formatters';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { getCallbackState, storeProviderToken } from './utils';
import { useAsyncFn } from 'react-use';

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
  const [{ loading, value, error }, req] = useAsyncFn(() => {
    let redirectUri = '';
    if (provider === GitProvider.GitLab) {
      redirectUri = gitlabOAuthRedirectURI();
    }

    return gitAuth.AuthorizeGitlab({
      redirectUri,
      code,
    });
  }, []);
  // const [res, loading, error, req] = useRequestState<AuthorizeGitlabResponse>();
  const linkResolver = useLinkResolver();
  const { setNotifications } = useNotifications();
  const { gitAuth } = useEnterpriseClient();
  const params = qs.parse(history.location.search);

  React.useEffect(() => {
    req();

    // if (provider === GitProvider.BitBucketServer) {
    //   req(
    //     gitAuth.AuthorizeBitbucketServer({
    //       code,
    //       state,
    //       redirectUri: bitbucketServerOAuthRedirectURI(),
    //     }),
    //   );
    // }

    // if (provider === GitProvider.AzureDevOps) {
    //   req(
    //     gitAuth.AuthorizeAzureDevOps({
    //       code,
    //       state,
    //       redirectUri: azureDevOpsOAuthRedirectURI(),
    //     }),
    //   );
    // }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [req]);

  React.useEffect(() => {
    if (!value) {
      return;
    }

    storeProviderToken(provider, value.token || '');

    const state = getCallbackState();

    if (state?.page || !params.error) {
      history.push(linkResolver(state?.page || ''));
      return;
    }
  }, [value, history, linkResolver, params.error, provider]);

  React.useEffect(() => {
    if (error) {
      setNotifications([{ message: { text: '' }, severity: 'error' }]);
    }
  }, [error, setNotifications]);

  return (
    <Page loading={loading} path={[{ label: 'OAuth Callback', url: '' }]}>
      <NotificationsWrapper>
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
      </NotificationsWrapper>
    </Page>
  );
}

export default styled(OAuthCallback).attrs({ className: OAuthCallback.name })``;
