import React, { FC } from "react";
import { BrowserRouter } from "react-router-dom";
import { muiTheme } from "./muiTheme";
import { MuiThemeProvider } from "@material-ui/core/styles";
import "@fortawesome/fontawesome-free/css/all.css";
import { createGlobalStyle, ThemeProvider } from "styled-components";
import theme from "weaveworks-ui-components/lib/theme";
import { Theme } from "weaveworks-ui-components";
import ProximaNova from "./fonts/proximanova-regular.woff";
import RobotoMono from "./fonts/roboto-mono-regular.woff";
import ResponsiveDrawer from "./components/ResponsiveDrawer";

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

  body {
    background: #e6e6e6;
    color: ${theme.textColor};
    font-family: ${theme.fontFamilies.regular};
    font-size: ${theme.fontSizes.normal};

    /* Layout - grow to at least viewport height */
    display: flex;
    min-height: 100%;
    flex-direction: column;
    margin: 0;
  }

  a {
    text-decoration: none;
  }
`;

const App: FC = () => {
  return (
    <BrowserRouter>
      <ThemeProvider theme={theme as Theme}>
        <MuiThemeProvider theme={muiTheme}>
          <GlobalStyle />
          <ResponsiveDrawer />
        </MuiThemeProvider>
      </ThemeProvider>
    </BrowserRouter>
  );
};

export default App;
