
import { act, render, RenderResult, screen } from '@testing-library/react';
import Applications from '../';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import { GitAuthProvider } from '../../../contexts/GitAuth';
import NotificationsProvider from '../../../contexts/Notifications/Provider';
import RequestContextProvider from '../../../contexts/Request';
import { muiTheme } from '../../../muiTheme';
import {
  ApplicationsClientMock,
  CoreClientMock,
  EnterpriseClientMock,
  withContext,
} from '../../../utils/test-utils';
import { MuiThemeProvider } from '@material-ui/core';
import {
  AppContextProvider,
  CoreClientContextProvider,
  Kind,
  theme,
  ThemeTypes,
} from '@weaveworks/weave-gitops';
import { QueryClient, QueryClientProvider } from 'react-query';
import { MemoryRouter } from 'react-router-dom';
import { ThemeProvider } from 'styled-components';

describe('Applications index test', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: CoreClientMock;
  let appsApi: ApplicationsClientMock;
  const appliedTheme = theme(ThemeTypes.Light);
  beforeEach(() => {
    window.matchMedia = jest.fn();
    //@ts-ignore
    window.matchMedia.mockReturnValue({ matches: false });
    api = new CoreClientMock();
    appsApi = new ApplicationsClientMock();
    wrap = withContext([
      [ThemeProvider, { theme: appliedTheme }],
      [
        MuiThemeProvider,
        { theme: muiTheme(appliedTheme.colors, ThemeTypes.Light) },
      ],
      [AppContextProvider],
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
      [GitAuthProvider, { api: appsApi }],
      [MemoryRouter],
      [NotificationsProvider],
    ]);
  });
  it('renders table rows', async () => {
    api.ListObjectsReturns = {
      [Kind.Kustomization]: {
        errors: [],
        objects: [
          {
            uid: 'uid1',
            payload: JSON.stringify({
              // maybe?
              apiVersion: 'kustomize.toolkit.fluxcd.io/v1beta2',
              kind: 'Kustomization',
              metadata: {
                namespace: 'my-ns',
                name: 'my-kustomization',
                uid: 'uid1',
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
      },
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
