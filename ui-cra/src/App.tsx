import '@fortawesome/fontawesome-free/css/all.css';
import { MuiThemeProvider } from '@material-ui/core/styles';
import {
  AppContextProvider,
  applicationsClient,
  coreClient,
  CoreClientContextProvider,
  FeatureFlagsContextProvider,
  theme,
} from '@weaveworks/weave-gitops';
import React, { FC } from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { BrowserRouter } from 'react-router-dom';
import { createGlobalStyle, ThemeProvider } from 'styled-components';
import Background from './assets/img/background.svg';
import ResponsiveDrawer from './components/ResponsiveDrawer';
import ProximaNova from './fonts/proximanova-regular.woff';
import RobotoMono from './fonts/roboto-mono-regular.woff';
import { muiTheme } from './muiTheme';

const GlobalStyle = createGlobalStyle`
  /* https://github.com/weaveworks/wkp-ui/pull/283#discussion_r339958886 */
  /* https://github.com/necolas/normalize.css/issues/694 */
  button,
  input,
  optgroup,
  select,
  textarea {
    font-family: inherit;
    font-size: 100%;
  }

  @font-face {
    font-family: 'proxima-nova';
    src: url(${ProximaNova})
  }

  @font-face {
    font-family: 'Roboto Mono';
    src: url(${RobotoMono})
  }

  html, body {
    height: 100%;
  }

  body {
    background: right bottom url(${Background}) no-repeat fixed ${theme.colors.neutral10}; 
    background-size: 100%;
    color: ${theme.colors.black};
    font-family: ${theme.fontFamilies.regular};
    font-size: ${theme.fontSizes.normal};
    /* Layout - grow to at least viewport height */
    display: flex;
    flex-direction: column;
    margin: 0;
  }

  a {
    text-decoration: none;
  }
`;

const queryClient = new QueryClient();

const App: FC = () => {
  return (
    <ThemeProvider theme={theme}>
      <MuiThemeProvider theme={muiTheme}>
        <BrowserRouter basename={process.env.PUBLIC_URL}>
          <QueryClientProvider client={queryClient}>
            <CoreClientContextProvider api={coreClient}>
              <GlobalStyle />
              <AppContextProvider applicationsClient={applicationsClient}>
                <FeatureFlagsContextProvider>
                  <ResponsiveDrawer />
                </FeatureFlagsContextProvider>
              </AppContextProvider>
            </CoreClientContextProvider>
          </QueryClientProvider>
        </BrowserRouter>
      </MuiThemeProvider>
    </ThemeProvider>
  );
};

export default App;
