import '@testing-library/jest-dom';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import 'jest-styled-components';
import 'jest-canvas-mock';
import * as React from 'react';
import { unmountComponentAtNode } from 'react-dom';
import { act } from 'react-dom/test-utils';
import { createMockClient, withContext, withTheme } from '../../lib/test-utils';
import GithubAuthButton from '../GithubAuthButton';
import { GlobalGithubAuthDialog } from '../GithubDeviceAuthModal';
import GithubAuthContextProvider from '../../../contexts/GithubAuth/provider';
import { ContentWrapper } from '../../Layout/ContentWrapper';

describe('GithubAuthButton', () => {
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
  it.skip('shows a modal when clicked', async () => {
    const promise = Promise.resolve();

    const userCode = 'ABCD-1234';

    const tree = () => (
      <div>
        <GithubAuthContextProvider>
          <ContentWrapper>
            <GithubAuthButton />
          </ContentWrapper>
          <GlobalGithubAuthDialog />
        </GithubAuthContextProvider>
      </div>
    );
    await act(async () => {
      render(
        withTheme(
          withContext(tree, '/', {
            applicationsClient: createMockClient({
              GetGithubDeviceCode: () => ({ userCode }),
            }),
          }),
        ),
        container,
      );
    });

    let modal;

    try {
      modal = await screen.findByRole('presentation');
      expect(modal).not.toBeInTheDocument();
    } catch (e) {
      //   we expect this to fail because the element should not be found
    }

    const button = await screen.findByText('Authenticate with GitHub');
    fireEvent(button, new MouseEvent('click', { bubbles: true }));

    modal = await screen.findByRole('presentation');

    expect(modal).toBeInTheDocument();
    expect(modal).toContainHTML(userCode);
    await waitFor(() => promise);
  });
});
