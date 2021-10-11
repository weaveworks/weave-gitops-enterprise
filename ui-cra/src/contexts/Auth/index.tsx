//@ts-nocheck
import _ from 'lodash';
import * as React from 'react';
import { ClustersService } from './../../capi-server/capi_server.pb';
import { GitProviderName } from '../../types/custom';
import { getProviderToken, storeProviderToken } from './../../utils/auth';

type AppState = {
  error: null | { fatal: boolean; message: string; detail?: string };
};

export type LinkResolver = (incoming: string) => string;

export function defaultLinkResolver(incoming: string): string {
  return incoming;
}

export type AuthContextType = {
  clustersService: typeof ClustersService;
  doAsyncError: (message: string, detail: string) => void;
  appState: AppState;
  linkResolver: LinkResolver;
  getProviderToken: typeof getProviderToken;
  storeProviderToken: typeof storeProviderToken;
};

export const AuthContext = React.createContext<AuthContextType>(
  null as AuthContextType,
);

export interface Props {
  clustersService: typeof ClustersService;
  linkResolver?: LinkResolver;
  children?: any;
}

// Due to the way the grpc-gateway typescript client is generated,
// we need to wrap each individual call to ensure the auth header gets
// injected into the underlying `fetch` requests.
// This saves us from having to rememeber to pass it as an arg in every request.
function wrapClient<T>(client: any, tokenGetter: () => string): T {
  const wrapped = {};

  _.each(client, (func, name) => {
    wrapped[name] = (payload, options: RequestInit = {}) => {
      const withToken: RequestInit = {
        ...options,
        headers: new Headers({
          ...(options.headers || {}),
          Authorization: `token ${tokenGetter()}`,
        }),
      };

      return func(payload, withToken);
    };
  });

  return wrapped as T;
}

export default function AuthContextProvider({
  clustersService,
  ...props
}: Props) {
  const [appState, setAppState] = React.useState({
    error: null,
  });

  React.useEffect(() => {
    // clear the error state on navigation
    setAppState({
      ...appState,
      error: null,
    });
    // adding appState as dependency leads to max depth exceeded
  }, [window.location]);

  const doAsyncError = (message: string, detail: string) => {
    console.error(message);
    setAppState({
      ...appState,
      error: { message, detail },
    });
  };

  const value: AuthContextType = {
    clustersService: wrapClient(clustersService, () =>
      getProviderToken(GitProviderName.GitHub),
    ),
    doAsyncError,
    appState,
    linkResolver: props.linkResolver || defaultLinkResolver,
    getProviderToken,
    storeProviderToken,
  };

  return <AuthContext.Provider {...props} value={value} />;
}
