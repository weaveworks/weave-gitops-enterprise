import { act, fireEvent, render, screen } from '@testing-library/react';
import { GitProvider } from '../../../api/applications/applications.pb';
import { GitAuthProvider } from '../../../contexts/GitAuth';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import { Routes } from '../../../utils/nav';
import {
  ApplicationsClientMock,
  defaultContexts,
  withContext,
} from '../../../utils/test-utils';
import RepoInputWithAuth from '../RepoInputWithAuth';
import { gitlabOAuthRedirectURI } from '../../../utils/formatters';

Object.assign(navigator, {
  clipboard: {
    writeText: () => {},
  },
});

describe('Gitlab Authenticate', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: ApplicationsClientMock;
  beforeEach(() => {
    let clipboardContents = '';

    Object.assign(navigator, {
      clipboard: {
        writeText: (text: string) => {
          clipboardContents = text;
          return Promise.resolve(text);
        },
        readText: () => Promise.resolve(clipboardContents),
      },
    });

    api = new ApplicationsClientMock();
    wrap = withContext([...defaultContexts(), [GitAuthProvider, { api }]]);
  });

  it('renders', async () => {
    await act(async () => {
      const c = wrap(
        <CallbackStateContextProvider
          callbackState={{
            page: Routes.AddApplication,
            state: { foo: 'bar' },
          }}
        >
          <RepoInputWithAuth onAuthClick={() => null} />
        </CallbackStateContextProvider>,
      );
      render(c);
    });
  });

  it('displays a button for GitLab auth', async () => {
    const repoUrl = 'https://gitlab.git.something/folder/name';
    const oauthUrl = 'https://gitlab.com/oauth/something';
    const capture = jest.fn();

    api.ParseRepoURLReturn = {
      name: 'somerepo',
      provider: GitProvider.GitLab,
      owner: 'someuser',
    };

    api.GetGitlabAuthURLReturn = {
      url: oauthUrl,
    };

    const onProviderChange = jest.fn();

    await act(async () => {
      const c = wrap(
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
      );
      render(c);
    });

    const button = (await (
      await screen.findByText('Authenticate with GitLab')
    ).closest('button')) as Element;
    expect(onProviderChange).toHaveBeenCalledWith(GitProvider.GitLab);
    fireEvent(button, new MouseEvent('click', { bubbles: true }));
    expect(capture).toHaveBeenCalledWith({
      redirectUri: gitlabOAuthRedirectURI(),
    });
  });
});
