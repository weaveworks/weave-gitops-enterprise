/// <reference types="react" />

declare module '@weaveworks/weave-gitops' {
  import _AppContextProvider from './contexts/AppContext';
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
  export declare const applicationsClient: typeof appsClient;
}
