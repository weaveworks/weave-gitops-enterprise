import { act, fireEvent, render, screen } from '@testing-library/react';
import { GitProvider } from '../../../api/gitauth/gitauth.pb';
import { GitAuthProvider } from '../../../contexts/GitAuth';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import { Routes } from '../../../utils/nav';
import {
  ApplicationsClientMock,
  defaultContexts,
  promisify,
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
    const repoUrl = 'https://gitlab.git.something/someuser/somerepo';
    const oauthUrl = 'https://gitlab.com/oauth/something';

    const capture = jest.fn();

    api.ParseRepoURLReturn = {
      name: 'somerepo',
      provider: GitProvider.GitLab,
      owner: 'someuser',
    };

    api.GetGitlabAuthURL = (req: any) => {
      capture(req);
      return promisify({ url: oauthUrl });
    };

    api.ValidateProviderTokenReturn = {
      valid: false,
    };

    const onProviderChange = jest.fn();
    const onAuthClick = jest.fn();

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
            onAuthClick={onAuthClick}
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
