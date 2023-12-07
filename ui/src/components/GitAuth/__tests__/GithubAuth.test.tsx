import { act, fireEvent, render, screen } from '@testing-library/react';
import {
  GetGithubDeviceCodeResponse,
  GitProvider,
} from '../../../api/gitauth/gitauth.pb';
import { EnterpriseClientContext } from '../../../contexts/API';
import {
  ApplicationsClientMock,
  defaultContexts,
  withContext,
} from '../../../utils/test-utils';
import { GithubDeviceAuthModal } from '../GithubDeviceAuthModal';
import { getProviderToken } from '../utils';

Object.assign(navigator, {
  clipboard: {
    writeText: () => {
      return;
    },
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
    wrap = withContext([
      ...defaultContexts(),
      [EnterpriseClientContext.Provider, { value: { gitAuth: api } }],
    ]);
  });

  it('renders the GithubAuth modal and user code', async () => {
    api.GetGithubDeviceCodeReturn = {
      userCode: 'D410-08FF',
      deviceCode: 'd725410cbe2431c5fa5dfa93736304db124412b6',
      validationUri: 'https://github.com/login/device',
      interval: 5,
    } as GetGithubDeviceCodeResponse;

    await act(async () => {
      const c = wrap(
        <GithubDeviceAuthModal
          onClose={() => {
            return;
          }}
          onSuccess={() => {
            return;
          }}
          open={true}
          repoName="config"
        />,
      );
      render(c);
    });
    expect(await screen.findByText('AUTHORIZE GITHUB ACCESS')).toBeTruthy();

    const ghCode = screen.getByTestId('github-code');
    expect(ghCode.textContent).toEqual(api.GetGithubDeviceCodeReturn.userCode);
    await act(async () => {
      const copyButton = await (
        await screen.findByTestId('github-code-container')
      ).querySelector('svg');
      fireEvent.click(copyButton as Element);
      await navigator.clipboard.readText().then(code => {
        expect(ghCode.textContent).toEqual(code);
      });
    });
  });

  it('stores a token', async () => {
    const accessToken = 'sometoken';
    api.GetGithubDeviceCodeReturn = {
      userCode: 'D410-08FF',
    };
    api.GetGithubAuthStatusReturn = {
      accessToken,
    };

    await act(async () => {
      const c = wrap(
        <GithubDeviceAuthModal
          onClose={() => {
            return;
          }}
          onSuccess={() => {
            return;
          }}
          open={true}
          repoName="config"
        />,
      );
      render(c);
    });

    const token = getProviderToken(GitProvider.GitHub);
    expect(token).toEqual(accessToken);
  });
});
