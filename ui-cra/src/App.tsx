import '@fortawesome/fontawesome-free/css/all.css';
import { MuiThemeProvider } from '@material-ui/core/styles';
import { FC } from 'react';
import {
  QueryCache,
  QueryClient,
  QueryClientConfig,
  QueryClientProvider,
} from 'react-query';
import { BrowserRouter } from 'react-router-dom';
import { createGlobalStyle, ThemeProvider } from 'styled-components';
import { muiTheme } from './muiTheme';

import { ProgressiveDeliveryService } from '@weaveworks/progressive-delivery';
import {
  AppContextProvider,
  applicationsClient,
  theme,
} from '@weaveworks/weave-gitops';
import { Pipelines } from './api/pipelines/pipelines.pb';
import bg from './assets/img/bg.svg';
import ResponsiveDrawer from './components/ResponsiveDrawer';
import { GithubAuthProvider } from './contexts/GithubAuth';
import { PipelinesProvider } from './contexts/Pipelines';
import { ProgressiveDeliveryProvider } from './contexts/ProgressiveDelivery';
import RequestContextProvider from './contexts/Request';
import ProximaNova from './fonts/proximanova-regular.woff';
import RobotoMono from './fonts/roboto-mono-regular.woff';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { ListConfigProvider, VersionProvider } from './contexts/ListConfig';

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
    background-color:${theme.colors.backGrey} !important;
  }

  body {
    background: right bottom no-repeat fixed; 
    background-image: url(${bg}), linear-gradient(to bottom, rgba(85, 105, 145, .1) 5%, rgba(85, 105, 145, .1), rgba(85, 105, 145, .25) 35%);
    background-size: 100%;
    color: ${theme.colors.black};
    font-family: ${theme.fontFamilies.regular};
    font-size: ${theme.fontSizes.medium};
    /* Layout - grow to at least viewport height */
    display: flex;
    flex-direction: column;
    margin: 0;
  }

  a {
    text-decoration: none;
    color:  ${theme.colors.primary10};
  }

  ::-webkit-scrollbar-track {
    margin-top: 5px;
    -webkit-box-shadow: transparent;
    -moz-box-shadow: transparent;
    background-color: transparent;
    border-radius: 5px;
  }

  ::-webkit-scrollbar{
    width: 5px;
    height: 5px;
    background-color: transparent;
    margin-top: 50px;
  }
 ::-webkit-scrollbar-thumb {
    background-color: ${props => props.theme.colors.neutral30};
    border-radius: 5px;
  }
 ::-webkit-scrollbar-thumb:hover {
    background-color: ${props => props.theme.colors.neutral30};
  }
  ::-webkit-scrollbar:hover{
    width: 7px;
    height:7px;
  }
`;

interface Error {
  code: number;
  message: string;
}
export const queryOptions: QueryClientConfig = {
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
  queryCache: new QueryCache({
    onError: error => {
      const err = error as Error;
      const { pathname, search } = window.location;
      const redirectUrl = encodeURIComponent(`${pathname}${search}`);
      const url = redirectUrl
        ? `/sign_in?redirect=${redirectUrl}`
        : `/sign_in?redirect=/`;
      if (err.code === 401 && !window.location.href.includes('/sign_in')) {
        window.location.href = url;
      }
    },
  }),
};
const queryClient = new QueryClient(queryOptions);

const App: FC = () => {
  return (
    <ThemeProvider theme={theme}>
      <MuiThemeProvider theme={muiTheme}>
        <RequestContextProvider fetch={window.fetch}>
          <QueryClientProvider client={queryClient}>
            <BrowserRouter basename={process.env.PUBLIC_URL}>
              <GlobalStyle />
              <ProgressiveDeliveryProvider api={ProgressiveDeliveryService}>
                <PipelinesProvider api={Pipelines}>
                  <GithubAuthProvider api={applicationsClient}>
                    <AppContextProvider applicationsClient={applicationsClient}>
                      <ListConfigProvider>
                        <VersionProvider>
                          <ResponsiveDrawer />
                          <ToastContainer
                            position="top-center"
                            autoClose={5000}
                            newestOnTop={false}
                          />
                        </VersionProvider>
                      </ListConfigProvider>
                    </AppContextProvider>
                  </GithubAuthProvider>
                </PipelinesProvider>
              </ProgressiveDeliveryProvider>
            </BrowserRouter>
          </QueryClientProvider>
        </RequestContextProvider>
      </MuiThemeProvider>
    </ThemeProvider>
  );
};

export default App;
