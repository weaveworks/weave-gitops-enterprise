import { applicationsClient } from '@weaveworks/weave-gitops';
import * as React from 'react';
import {
  clearCallbackState,
  getCallbackState,
  getProviderToken,
  storeCallbackState,
  storeProviderToken,
} from './utils';

export type AppContextType = {
  applicationsClient: typeof applicationsClient;
  getProviderToken: typeof getProviderToken;
  storeProviderToken: typeof storeProviderToken;
  getCallbackState: typeof getCallbackState;
  storeCallbackState: typeof storeCallbackState;
  clearCallbackState: typeof clearCallbackState;
};

export const AppContext = React.createContext<AppContextType | null>(null);

export interface AppProps {
  applicationsClient: typeof applicationsClient;
  children?: any;
}

export default function AppContextProvider({
  applicationsClient,
  ...props
}: AppProps) {
  const value: AppContextType = {
    applicationsClient,
    getProviderToken,
    storeProviderToken,
    storeCallbackState,
    getCallbackState,
    clearCallbackState,
  };

  return <AppContext.Provider {...props} value={value} />;
}
