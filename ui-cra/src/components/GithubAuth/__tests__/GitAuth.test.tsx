import { act, fireEvent, render, screen } from '@testing-library/react';
import { GithubDeviceAuthModal } from '..';
import { GithubAuthProvider } from '../../../contexts/GitAuth';
import {
  ApplicationsClientMock,
  defaultContexts,
  withContext,
} from '../../../utils/test-utils';

Object.assign(navigator, {
  clipboard: {
    writeText: () => {},
  },
});

describe('Github Authenticate', () => {
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
    wrap = withContext([...defaultContexts(), [GithubAuthProvider, { api }]]);
  });

  it('render gitAuth modal', async () => {
    api.GetGithubDeviceCodeReturn = {
      userCode: 'D410-08FF',
      deviceCode: 'd725410cbe2431c5fa5dfa93736304db124412b6',
      validationURI: 'https://github.com/login/device',
      interval: 5,
    };

    await act(async () => {
      const c = wrap(
        <GithubDeviceAuthModal
          onClose={() => {}}
          onSuccess={() => {}}
          open={true}
          repoName="config"
        />,
      );
      render(c);
    });
    expect(await screen.findByText('Authenticate with Github')).toBeTruthy();

    const ghCode = screen.getByTestId('github-code');
    expect(ghCode.textContent).toEqual(api.GetGithubDeviceCodeReturn.userCode);
    await act(async () => {
      fireEvent.click(ghCode as Element);
      await navigator.clipboard.readText().then(code => {
        expect(ghCode.textContent).toEqual(code);
      });
    });
  });
});
