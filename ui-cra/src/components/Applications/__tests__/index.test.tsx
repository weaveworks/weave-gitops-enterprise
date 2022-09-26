import Applications from '../';

import { MuiThemeProvider } from '@material-ui/core';
import { act, render, RenderResult, screen } from '@testing-library/react';
import { CoreClientContextProvider, theme } from '@weaveworks/weave-gitops';
import { QueryClient, QueryClientProvider } from 'react-query';
import { MemoryRouter } from 'react-router-dom';
import { ThemeProvider } from 'styled-components';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import NotificationsProvider from '../../../contexts/Notifications/Provider';
import RequestContextProvider from '../../../contexts/Request';
import { muiTheme } from '../../../muiTheme';
import {
  CoreClientMock,
  EnterpriseClientMock,
  withContext,
} from '../../../utils/test-utils';

describe('Applications index test', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: CoreClientMock;

  beforeEach(() => {
    api = new CoreClientMock();
    wrap = withContext([
      [ThemeProvider, { theme: theme }],
      [MuiThemeProvider, { theme: muiTheme }],
      [
        RequestContextProvider,
        { fetch: () => new Promise(accept => accept(null)) },
      ],
      [QueryClientProvider, { client: new QueryClient() }],
      [
        EnterpriseClientProvider,
        {
          api: new EnterpriseClientMock(),
        },
      ],
      [CoreClientContextProvider, { api }],
      [MemoryRouter],
      [NotificationsProvider],
    ]);
  });
  it('renders table rows', async () => {
    api.ListObjectsReturns = {
      objects: [
        {
          payload: JSON.stringify({
            // maybe?
            apiVersion: 'kustomize.toolkit.fluxcd.io/v1beta2',
            kind: 'Kustomization',
            metadata: {
              namespace: 'my-ns',
              name: 'my-kustomization',
            },
            spec: {
              path: './',
              interval: {},
              sourceRef: {},
            },
            status: {
              conditions: [],
              lastAppliedRevision: '',
              lastAttemptedRevision: '',
              inventory: [],
            },
          }),
          clusterName: 'my-cluster',
        },
      ],
    };

    await act(async () => {
      const c = wrap(<Applications />);
      render(c);
    });

    expect(await screen.findByText('my-kustomization')).toBeTruthy();
  });

  describe('snapshots', () => {
    it('loading', async () => {
      await act(async () => {
        const c = wrap(<Applications />);
        const result = render(c);

        expect(result.container).toMatchSnapshot();
      });
    });
    it('success', async () => {
      let result: RenderResult;
      await act(async () => {
        const c = wrap(<Applications />);
        result = await render(c);
      });

      //   @ts-ignore
      expect(result.container).toMatchSnapshot();
    });
  });
});
