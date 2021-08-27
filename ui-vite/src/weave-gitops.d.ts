/// <reference types="react" />

declare module '@weaveworks/weave-gitops' {
  // import _AppContextProvider from './contexts/AppContext';
  // import { Applications as appsClient } from './lib/api/applications/applications.pb';
  export const theme: import('styled-components').DefaultTheme;
  export const AppContextProvider: import('styled-components').StyledComponent<
    ({ applicationsClient }: { applicationsClient?: string }) => JSX.Element,
    import('styled-components').DefaultTheme,
    {},
    never
  >;
  export const Applications: import('styled-components').StyledComponent<
    ({ className }: { className?: string }) => JSX.Element,
    import('styled-components').DefaultTheme,
    {},
    never
  >;
  export const ApplicationDetail: import('styled-components').StyledComponent<
    ({ className, name }: { className?: string; name: string }) => JSX.Element,
    import('styled-components').DefaultTheme,
    {},
    never
  >;
  export const applicationsClient: string;
}
