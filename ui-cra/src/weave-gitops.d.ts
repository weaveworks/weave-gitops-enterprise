/// <reference types="react" />

declare module '@weaveworks/weave-gitops' {
  import _AppContextProvider from './contexts/AppContext';
  import _useApplications from './hooks/applications';
  import _LoadingPage from './components/LoadingPage';
  import { Applications as appsClient } from './lib/api/applications/applications.pb';
  export declare const theme: import('styled-components').DefaultTheme;
  export declare const AppContextProvider: typeof _AppContextProvider;
  export declare const Applications: import('styled-components').StyledComponent<
    ({ className }: { className?: string }) => JSX.Element,
    import('styled-components').DefaultTheme,
    {},
    never
  >;
  export declare const ApplicationDetail: import('styled-components').StyledComponent<
    ({ className, name }: { className?: string; name: string }) => JSX.Element,
    import('styled-components').DefaultTheme,
    {},
    never
  >;
  export declare const ApplicationAdd: import('styled-components').StyledComponent<
    ({ className }: { className?: string }) => JSX.Element,
    import('styled-components').DefaultTheme,
    {},
    never
  >;
  export declare const GithubDeviceAuthModal: import('styled-components').StyledComponent<
    ({
      className,
      open,
      onSuccess,
      onClose,
      repoName,
    }: {
      className?: string;
      open: boolean;
      onSuccess: (token: string) => void;
      onClose: () => void;
      repoName: string;
    }) => JSX.Element,
    import('styled-components').DefaultTheme,
    {},
    never
  >;
  export declare const applicationsClient: typeof appsClient;
  export declare const useApplications: typeof _useApplications;
  export declare const LoadingPage: typeof _LoadingPage;
  export declare const getProviderToken: (string) => string;
  export declare const isUnauthenticated: (int) => boolean;
}
