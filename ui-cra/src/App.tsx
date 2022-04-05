import React, { FC } from 'react';
import { BrowserRouter } from 'react-router-dom';
import { muiTheme } from './muiTheme';
import { MuiThemeProvider } from '@material-ui/core/styles';
import '@fortawesome/fontawesome-free/css/all.css';
import { createGlobalStyle, ThemeProvider } from 'styled-components';
import {
  AuthContextProvider,
  FeatureFlagsContextProvider,
  theme,
} from '@weaveworks/weave-gitops';
import ProximaNova from './fonts/proximanova-regular.woff';
import RobotoMono from './fonts/roboto-mono-regular.woff';
import Background from './assets/img/background.svg';
import ResponsiveDrawer from './components/ResponsiveDrawer';

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
    background: url(${Background}) no-repeat;
    background-color: ${theme.colors.neutral10};
    background-position: right bottom;
    background-attachment:fixed;
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

const App: FC = () => {
  return (
    <ThemeProvider theme={theme}>
      <MuiThemeProvider theme={muiTheme}>
        <BrowserRouter basename={process.env.PUBLIC_URL}>
          <FeatureFlagsContextProvider>
            <AuthContextProvider>
              <GlobalStyle />
              <ResponsiveDrawer />
            </AuthContextProvider>
          </FeatureFlagsContextProvider>
        </BrowserRouter>
      </MuiThemeProvider>
    </ThemeProvider>
  );
};

export default App;
