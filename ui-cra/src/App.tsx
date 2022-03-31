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

  .text-center {
    text-align: center!important;
  }
  .alert-danger {
      background-color: #f8d7da;
      border-color: #f5c2c7;
      color: #842029;
  }
  .alert {
      border: 1px solid transparent;
      border-radius: 0.25rem;
      margin-bottom: 1rem;
      padding: 1rem;
      display: flex;
      align-items: center;
      justify-content: center;
  }
  .retry{
    margin-left: 4px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .flex-center{
    display: flex;
    align-items:center;
  }
  .severity-icon{
    font-size: 14px;
    margin-right:4px;
  }
  .severity-low{
    color:#DFD41B
  }
  .severity-medium{
    color:#FF7000
  }
  .severity-high{
    color:#E2423B
  }
  .flex-center{
    display:flex;
    lign-items: center;
    justify-content: center;
  }
  .flex-start{
    display:flex;
    align-items: center;
    justify-content: start;
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
