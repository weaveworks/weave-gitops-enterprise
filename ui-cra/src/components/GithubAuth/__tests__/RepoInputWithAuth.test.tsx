import { fireEvent, render, screen } from '@testing-library/react';
import { ParseRepoURLResponse } from '@weaveworks/weave-gitops';
import 'jest-styled-components';
import * as React from 'react';
import { unmountComponentAtNode } from 'react-dom';
import { act } from 'react-dom/test-utils';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import { GitProvider } from '../utils';
import { gitlabOAuthRedirectURI } from '../../../utils/formatters';
import { Routes } from '../../../utils/nav';
// check how we do this in EE
import { createMockClient, withContext, withTheme } from '../../lib/test-utils';
import RepoInputWithAuth from '../RepoInputWithAuth';

describe('RepoInputWithAuth', () => {
  let container: any = null;
  beforeEach(() => {
    // setup a DOM element as a render target
    container = document.createElement('div');
    document.body.appendChild(container);
  });

  afterEach(() => {
    // cleanup on exiting
    unmountComponentAtNode(container);
    container.remove();
    container = null;
  });
  describe('snapshots', () => {
    it('renders', async () => {
      await act(async () => {
        render(
          withTheme(
            withContext(
              <CallbackStateContextProvider
                callbackState={{
                  page: Routes.AddApplication,
                  state: { foo: 'bar' },
                }}
              >
                <RepoInputWithAuth onAuthClick={() => null} />
              </CallbackStateContextProvider>,
              '/',
              {},
            ),
          ),
        );
      });
    });
    it('requests URL parsing', async () => {
      const url = 'git@github.com:someuser/somerepo.git';
      const c = {
        ParseRepoURL: jest.fn(),
      };

      const onProviderChange = jest.fn();

      await act(async () => {
        render(
          withTheme(
            withContext(
              <CallbackStateContextProvider
                callbackState={{
                  page: Routes.AddApplication,
                  state: { foo: 'bar' },
                }}
              >
                <RepoInputWithAuth
                  value={url}
                  onAuthClick={() => null}
                  onProviderChange={onProviderChange}
                />
              </CallbackStateContextProvider>,
              '/',
              { applicationsClient: createMockClient(c) },
            ),
          ),
        );
      });

      expect(c.ParseRepoURL).toBeCalledTimes(1);
      expect(c.ParseRepoURL).toBeCalledWith({ url });
    });
    it('displays a button for GitHub auth', async () => {
      const url = 'git@github.com:someuser/somerepo.git';
      const c = {
        ParseRepoURL: (): ParseRepoURLResponse => ({
          name: 'somerepo',
          provider: GitProvider.GitHub,
          owner: 'someuser',
        }),
      };

      const onAuthClick = jest.fn();
      const onProviderChange = jest.fn();

      await act(async () => {
        render(
          withTheme(
            withContext(
              <CallbackStateContextProvider
                callbackState={{
                  page: Routes.AddApplication,
                  state: { foo: 'bar' },
                }}
              >
                <RepoInputWithAuth
                  value={url}
                  onAuthClick={onAuthClick}
                  onProviderChange={onProviderChange}
                />
              </CallbackStateContextProvider>,

              '/',
              { applicationsClient: createMockClient(c) },
            ),
          ),
        );
      });

      const button = await (
        await screen.findByText('Authenticate with GitHub')
      ).closest('button');
      expect(onProviderChange).toHaveBeenCalledWith(GitProvider.GitHub);
      fireEvent(button, new MouseEvent('click', { bubbles: true }));
      expect(onAuthClick).toHaveBeenCalledWith(GitProvider.GitHub);
    });
    it('displays a button for GitLab auth', async () => {
      const repoUrl = 'git@gitlab.com:someuser/somerepo.git';
      const oauthUrl = 'https://gitlab.com/oauth/something';
      const capture = jest.fn();
      const c = {
        ParseRepoURL: (): ParseRepoURLResponse => ({
          name: 'somerepo',
          provider: GitProvider.GitLab,
          owner: 'someuser',
        }),
        GetGitlabAuthURL: req => {
          capture(req);
          return { url: oauthUrl };
        },
      };

      const onProviderChange = jest.fn();

      await act(async () => {
        render(
          withTheme(
            withContext(
              <CallbackStateContextProvider
                callbackState={{
                  page: Routes.AddApplication,
                  state: { foo: 'bar' },
                }}
              >
                <RepoInputWithAuth
                  value={repoUrl}
                  onProviderChange={onProviderChange}
                  onAuthClick={() => null}
                />
              </CallbackStateContextProvider>,
              '/',
              { applicationsClient: createMockClient(c) },
            ),
          ),
        );
      });

      const button = await (
        await screen.findByText('Authenticate with GitLab')
      ).closest('button');
      expect(onProviderChange).toHaveBeenCalledWith(GitProvider.GitLab);
      fireEvent(button, new MouseEvent('click', { bubbles: true }));
      expect(capture).toHaveBeenCalledWith({
        redirectUri: gitlabOAuthRedirectURI(),
      });
    });
  });
});
