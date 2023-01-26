import { CircularProgress } from '@material-ui/core';
import { Alert, AlertTitle } from '@material-ui/lab';
import {
  Flex,
  useLinkResolver,
  useRequestState,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';
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
import { gitlabOAuthRedirectURI } from '../../utils/formatters';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { getCallbackState, storeProviderToken } from './utils';

type Props = {
  code: string;
  provider: GitProvider;
};

type BitBucketErrorParams = {
  error?: string[];
  error_description?: string[];
};

function OAuthCallback({ code, provider }: Props) {
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [code, gitAuthClient, provider]);

  React.useEffect(() => {
    if (!res) {
      return;
    }

    storeProviderToken(GitProvider.GitLab, res.token || '');

    const state = getCallbackState();

    if (state?.page || !params.error) {
      history.push(linkResolver(state?.page || ''));
      return;
    }
  }, [res, history, linkResolver]);

  React.useEffect(() => {
    if (error) {
      setNotifications([{ message: { text: '' }, severity: 'error' }]);
    }
  }, [error, setNotifications]);

  return (
    <PageTemplate path={[{ label: 'OAuth Callback', url: '' }]}>
      <ContentWrapper loading={loading}>
        <Flex wide align center>
          {loading && <CircularProgress />}
          {(params as BitBucketErrorParams)?.error && (
            <Alert severity="error">
              <AlertTitle>
                Oauth Error:{' '}
                {_.isArray(params?.error)
                  ? params?.error.join(', ')
                  : params?.error}
              </AlertTitle>
              {(params as BitBucketErrorParams)?.error_description?.join('\n')}
            </Alert>
          )}
        </Flex>
      </ContentWrapper>
    </PageTemplate>
  );
}

export default styled(OAuthCallback).attrs({ className: OAuthCallback.name })``;
