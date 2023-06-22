import '@fortawesome/fontawesome-free/css/all.css';
import { MuiThemeProvider } from '@material-ui/core/styles';
import { ProgressiveDeliveryService } from '@weaveworks/progressive-delivery';
import {
  AppContext,
  AppContextProvider,
  AuthCheck,
  AuthContextProvider,
  CoreClientContextProvider,
  LinkResolverProvider,
  Pendo,
  SignIn,
  ThemeTypes,
  coreClient,
  theme,
} from '@weaveworks/weave-gitops';
import React, { ReactNode } from 'react';
import {
  QueryCache,
  QueryClient,
  QueryClientConfig,
  QueryClientProvider,
} from 'react-query';
import { BrowserRouter, Route, Switch } from 'react-router-dom';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { ThemeProvider, createGlobalStyle } from 'styled-components';
import { Pipelines } from './api/pipelines/pipelines.pb';
import { Query } from './api/query/query.pb';
import { Terraform } from './api/terraform/terraform.pb';
import bgDark from './assets/img/bg-dark.png';
import bg from './assets/img/bg.svg';
import { ClustersService } from './cluster-services/cluster_services.pb';
import App from './components/Layout/App';
import MemoizedHelpLinkWrapper from './components/Layout/HelpLinkWrapper';
import Compose from './components/ProvidersCompose';
import EnterpriseClientProvider from './contexts/EnterpriseClient/Provider';
import { GitAuthProvider } from './contexts/GitAuth/index';
import NotificationsProvider from './contexts/Notifications/Provider';
import { PipelinesProvider } from './contexts/Pipelines';
import { ProgressiveDeliveryProvider } from './contexts/ProgressiveDelivery';
import QueryServiceProvider from './contexts/QueryService';
import RequestContextProvider from './contexts/Request';
import { TerraformProvider } from './contexts/Terraform';
import ProximaNova from './fonts/proximanova-regular.woff';
import RobotoMono from './fonts/roboto-mono-regular.woff';
import { muiTheme } from './muiTheme';
import { resolver } from './utils/link-resolver';

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

  html, body, #root {
    height: 100%;
    margin: 0;
  }

  body {
    background: right bottom no-repeat fixed; 
    background-image: ${props =>
      props.theme.mode === ThemeTypes.Dark
        ? `url(${bgDark});`
        : `url(${bg}), linear-gradient(to bottom, rgba(85, 105, 145, .1) 5%, rgba(85, 105, 145, .1), rgba(85, 105, 145, .25) 35%);`}
    background-size: 100%;
    background-color: ${props =>
      props.theme.mode === ThemeTypes.Dark
        ? props.theme.colors.neutralGray
        : 'transparent'};
    color: ${props => props.theme.colors.black};
    font-family: ${props => props.theme.fontFamilies.regular};
    font-size: ${props => props.theme.fontSizes.medium};
    //we can not override Autocomplete in Mui createTheme without updating our Mui version. 
    .policies-input {
      border: 1px solid ${props => props.theme.colors.neutral20};
    }
    .MuiAutocomplete-inputRoot {
      &.MuiInputBase-root {
        padding: 0;
      }
    }
    .MuiAutocomplete-groupLabel {
      background: ${props => props.theme.colors.neutralGray};
    }

    a {
      text-decoration: none;
      color:  ${props => props.theme.colors.primary10};
    }
    .MuiFormControl-root {
      min-width: 0px;
    }
   
  }
  ::-webkit-scrollbar-track {
    margin-top: 5px;
    box-shadow: transparent;
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
  #root {
    /* Layout - grow to at least viewport height */
    display: flex;
    flex-direction: column;
    justify-content: center;
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

const StylesProvider = ({ children }: { children: ReactNode }) => {
  const { settings } = React.useContext(AppContext);
  const mode = settings.theme;
  const appliedTheme = theme(mode);
  return (
    <ThemeProvider theme={appliedTheme}>
      <MuiThemeProvider theme={muiTheme(appliedTheme.colors, mode)}>
        <GlobalStyle />
        {children}
      </MuiThemeProvider>
    </ThemeProvider>
  );
};

const AppContainer = () => {
  return (
    <RequestContextProvider fetch={window.fetch}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter basename={process.env.PUBLIC_URL}>
          <ProgressiveDeliveryProvider api={ProgressiveDeliveryService}>
            <PipelinesProvider api={Pipelines}>
              <GitAuthProvider>
                <AppContextProvider footer={<MemoizedHelpLinkWrapper />}>
                  <StylesProvider>
                    <AuthContextProvider>
                      <EnterpriseClientProvider api={ClustersService}>
                        <CoreClientContextProvider api={coreClient}>
                          <TerraformProvider api={Terraform}>
                            <LinkResolverProvider resolver={resolver}>
                              <Pendo
                                defaultTelemetryFlag="true"
                                tier="enterprise"
                                version={process.env.REACT_APP_VERSION}
                              />
                              <QueryServiceProvider api={Query}>
                                <Compose components={[NotificationsProvider]}>
                                  <Switch>
                                    <Route
                                      component={() => <SignIn />}
                                      exact={true}
                                      path="/sign_in"
                                    />
                                    <Route path="*">
                                      {/* Check we've got a logged in user otherwise redirect back to signin */}
                                      <AuthCheck>
                                        <App />
                                      </AuthCheck>
                                    </Route>
                                  </Switch>
                                  <ToastContainer
                                    position="top-center"
                                    autoClose={5000}
                                    newestOnTop={false}
                                  />
                                </Compose>
                              </QueryServiceProvider>
                            </LinkResolverProvider>
                          </TerraformProvider>
                        </CoreClientContextProvider>
                      </EnterpriseClientProvider>
                    </AuthContextProvider>
                  </StylesProvider>
                </AppContextProvider>
              </GitAuthProvider>
            </PipelinesProvider>
          </ProgressiveDeliveryProvider>
        </BrowserRouter>
      </QueryClientProvider>
    </RequestContextProvider>
  );
};

export default AppContainer;
