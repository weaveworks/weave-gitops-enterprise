import { act, fireEvent, render, screen } from '@testing-library/react';
import { GitProvider } from '../../../api/gitauth/gitauth.pb';
import { GitAuthProvider } from '../../../contexts/GitAuth';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import { gitlabOAuthRedirectURI } from '../../../utils/formatters';
import { Routes } from '../../../utils/nav';
import {
  ApplicationsClientMock,
  defaultContexts,
  promisify,
  withContext,
} from '../../../utils/test-utils';
import { RepoInputWithAuth } from '../RepoInputWithAuth';

Object.assign(navigator, {
  clipboard: {
    writeText: () => {},
  },
});

describe('Gitlab Authenticate', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: ApplicationsClientMock;

  const gitRepoUrl = JSON.stringify({
    value: 'https://gitlab.git.dev.weave.works/test',
  });

  const onProviderChange = jest.fn();
  const onAuthClick = jest.fn();
  const setFormData = jest.fn();

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
          <RepoInputWithAuth
            value={gitRepoUrl}
            onProviderChange={onProviderChange}
            onAuthClick={onAuthClick}
            label=""
            formData=""
            setFormData={setFormData}
            loading={false}
          />
        </CallbackStateContextProvider>,
      );
      render(c);
    });
  });

  it('displays a button for GitLab auth', async () => {
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

    await act(async () => {
      const c = wrap(
        <CallbackStateContextProvider
          callbackState={{
            page: Routes.AddApplication,
            state: { foo: 'bar' },
          }}
        >
          <RepoInputWithAuth
            value={gitRepoUrl}
            onProviderChange={onProviderChange}
            onAuthClick={onAuthClick}
            label=""
            formData=""
            setFormData={setFormData}
            loading={false}
          />
        </CallbackStateContextProvider>,
      );
      render(c);
    });

    const button = (await (
      await screen.findByText('AUTHENTICATE WITH GITLAB')
    ).closest('button')) as Element;
    expect(onProviderChange).toHaveBeenCalledWith(GitProvider.GitLab);
    fireEvent(button, new MouseEvent('click', { bubbles: true }));
    expect(capture).toHaveBeenCalledWith({
      redirectUri: gitlabOAuthRedirectURI(),
    });
  });
});
