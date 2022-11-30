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
import { getCallbackState, GitProvider, storeProviderToken } from './utils';
import { gitlabOAuthRedirectURI } from '../../utils/formatters';
import { ContentWrapper } from '../Layout/ContentWrapper';
import useNotifications from '../../contexts/Notifications';

type Props = {
  code: string;
  provider: GitProvider;
};

function OAuthCallback({ code, provider }: Props) {
  const history = useHistory();
  const { applicationsClient } = React.useContext(AppContext);
  const [res, loading, error, req] = useRequestState<AuthorizeGitlabResponse>();
  const linkResolver = useLinkResolver();
  const { setNotifications } = useNotifications();

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
  }, [
    code,
    applicationsClient,
    provider,
    // req causes an infinite loop
  ]);

  React.useEffect(() => {
    if (!res) {
      return;
    }

    storeProviderToken(GitProvider.GitLab, res.token || '');

    const state = getCallbackState();

    if (state?.page) {
      history.push(linkResolver(state.page));
      return;
    }
  }, [res, getCallbackState, history, linkResolver, storeProviderToken]);

  React.useEffect(() => {
    if (error) {
      setNotifications([{ message: { text: '' }, severity: 'error' }]);
    }
  }, [error, setNotifications]);

  return (
    <ContentWrapper loading={loading}>
      <Flex wide align center>
        {loading && <CircularProgress />}
      </Flex>
    </ContentWrapper>
  );
}

export default styled(OAuthCallback).attrs({ className: OAuthCallback.name })``;
