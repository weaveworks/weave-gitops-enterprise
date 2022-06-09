import Box from '@material-ui/core/Box';
import CssBaseline from '@material-ui/core/CssBaseline';
import Drawer from '@material-ui/core/Drawer';
import Hidden from '@material-ui/core/Hidden';
import IconButton from '@material-ui/core/IconButton';
import {
  createStyles,
  makeStyles,
  Theme,
  useTheme,
} from '@material-ui/core/styles';
import {
  AuthCheck,
  AuthContextProvider,
  coreClient,
  CoreClientContextProvider,
  OAuthCallback,
  SignIn,
  theme as weaveTheme,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import qs from 'query-string';
import React from 'react';
import Lottie from 'react-lottie-player';
import { Route, Switch } from 'react-router-dom';
import styled from 'styled-components';
import error404 from '../assets/img/error404.json';
import { ReactComponent as MenuIcon } from '../assets/img/menu-burger.svg';
import { ClustersService } from '../cluster-services/cluster_services.pb';
import ClustersProvider from '../contexts/Clusters/Provider';
import EnterpriseClientProvider from '../contexts/EnterpriseClient/Provider';
import NotificationsProvider from '../contexts/Notifications/Provider';
import ProgressiveDelivery from '../contexts/ProgressiveDelivery';
import TemplatesProvider from '../contexts/Templates/Provider';
import VersionsProvider from '../contexts/Versions/Provider';
import WGApplicationsDashboard from './Applications';
import WGApplicationsBucket from './Applications/Bucket';
import WGApplicationsFluxRuntime from './Applications/FluxRuntime';
import WGApplicationsGitRepository from './Applications/GitRepository';
import WGApplicationsHelmChart from './Applications/HelmChart';
import WGApplicationsHelmRelease from './Applications/HelmRelease';
import WGApplicationsHelmRepository from './Applications/HelmRepository';
import WGApplicationsKustomization from './Applications/Kustomization';
import WGApplicationsSources from './Applications/Sources';
import MCCP from './Clusters';
import AddClusterWithCredentials from './Clusters/Create';
import { ContentWrapper } from './Layout/ContentWrapper';
import { PageTemplate } from './Layout/PageTemplate';
import { SectionHeader } from './Layout/SectionHeader';
import { Navigation } from './Navigation';
import Policies from './Policies';
import PolicyDetails from './Policies/PolicyDetails';
import PoliciesViolations from './PolicyViolations';
import PolicyViolationDetails from './PolicyViolations/ViolationDetails';
import Canaries from './ProgressiveDelivery/Canaries';
import Compose from './ProvidersCompose';
import TemplatesDashboard from './Templates';

const GITLAB_OAUTH_CALLBACK = '/oauth/gitlab';
const POLICIES = '/policies';

function withSearchParams(Cmp: any) {
  return ({ location: { search }, ...rest }: any) => {
    const params = qs.parse(search);

    return <Cmp {...rest} {...params} />;
  };
}

const drawerWidth = 220;

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      display: 'flex',
    },
    drawer: {
      [theme.breakpoints.up('sm')]: {
        width: drawerWidth,
        flexShrink: 0,
      },
    },
    appBar: {
      [theme.breakpoints.up('sm')]: {
        width: `calc(100% - ${drawerWidth}px)`,
        marginLeft: drawerWidth,
      },
      backgroundColor: weaveTheme.colors.primary,
      boxShadow: 'none',
    },
    menuButton: {
      [theme.breakpoints.up('sm')]: {
        display: 'none',
      },
      marginLeft: 0,
    },
    menuButtonBox: {
      height: '80px',
      background: weaveTheme.colors.primary,
      display: 'flex',
      justifyContent: 'center',
      position: 'sticky',
      top: '0',
      zIndex: 2,
    },
    toolbar: theme.mixins.toolbar,
    drawerPaper: {
      overflowY: 'hidden',
      width: drawerWidth,
      border: 'none',
      background: weaveTheme.colors.neutral10,
    },
    content: {
      flexGrow: 1,
    },
  }),
);

const SignInWrapper = styled.div`
  height: 100vh;
  .MuiAlert-root {
    width: 470px;
  }
  .MuiDivider-root {
    width: 290px;
  }
  form div:nth-child(1),
  form div:nth-child(2) {
    padding-right: 10px;
  }
  .MuiInputBase-root {
    flex-grow: 0;
  }
  .MuiIconButton-root {
    padding-right: 0;
  }
`;

const CoreWrapper = styled.div`
  div[class*='FilterDialog__SlideContainer'] {
    overflow: hidden;
  }
  .MuiFormControl-root {
    min-width: 0px;
  }
  div[class*='ReconciliationGraph'] {
    svg {
      min-height: 600px;
    }
    .MuiSlider-root.MuiSlider-vertical {
      height: 200px;
    }
  }
  .MuiButton-root {
    margin-right: 0;
  }
  max-width: calc(100vw - 220px);
`;

const Page404 = () => (
  <PageTemplate documentTitle="WeGO · NotFound">
    <SectionHeader />
    <ContentWrapper>
      <Lottie
        loop
        animationData={error404}
        play
        style={{ width: '100%', height: 650 }}
      />
    </ContentWrapper>
  </PageTemplate>
);

const App = () => {
  const classes = useStyles();
  const theme = useTheme();
  const [mobileOpen, setMobileOpen] = React.useState(false);

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  return (
    <Compose
      components={[
        NotificationsProvider,
        TemplatesProvider,
        ClustersProvider,
        VersionsProvider,
        ProgressiveDelivery,
      ]}
    >
      <div className={classes.root}>
        <CssBaseline />
        <Box className={classes.menuButtonBox}>
          <IconButton
            color="inherit"
            aria-label="open drawer"
            edge="start"
            onClick={handleDrawerToggle}
            className={classes.menuButton}
          >
            <MenuIcon />
          </IconButton>
        </Box>
        <nav className={classes.drawer} aria-label="mailbox folders">
          <Hidden smUp implementation="css">
            <Drawer
              variant="temporary"
              anchor={theme.direction === 'rtl' ? 'right' : 'left'}
              open={mobileOpen}
              onClose={handleDrawerToggle}
              classes={{
                paper: classes.drawerPaper,
              }}
              ModalProps={{
                keepMounted: true, // Better open performance on mobile.
              }}
            >
              <div className={classes.toolbar}>
                <Navigation />
              </div>
            </Drawer>
          </Hidden>
          <Hidden xsDown implementation="css">
            <Drawer
              classes={{
                paper: classes.drawerPaper,
              }}
              variant="permanent"
              open
            >
              <div className={classes.toolbar}>
                <Navigation />
              </div>
            </Drawer>
          </Hidden>
        </nav>
        <main className={classes.content}>
          <Switch>
            <Route component={MCCP} exact path={['/', '/clusters']} />
            <Route component={MCCP} exact path="/clusters/delete" />
            <Route
              component={AddClusterWithCredentials}
              exact
              path="/clusters/templates/:templateName/create"
            />
            <Route
              component={TemplatesDashboard}
              exact
              path="/clusters/templates"
            />
            <Route
              component={PoliciesViolations}
              exact
              path="/clusters/violations"
            />
            <Route
              component={PolicyViolationDetails}
              exact
              path="/clusters/violations/:id"
            />
            <Route
              component={() => (
                <CoreWrapper>
                  <WGApplicationsDashboard />
                </CoreWrapper>
              )}
              exact
              path={V2Routes.Automations}
            />
            <Route
              component={() => (
                <CoreWrapper>
                  <WGApplicationsSources />
                </CoreWrapper>
              )}
              exact
              path={V2Routes.Sources}
            />
            <Route
              component={withSearchParams((props: any) => (
                <CoreWrapper>
                  <WGApplicationsKustomization {...props} />
                </CoreWrapper>
              ))}
              path={V2Routes.Kustomization}
            />
            <Route
              component={withSearchParams((props: any) => (
                <CoreWrapper>
                  <WGApplicationsGitRepository {...props} />
                </CoreWrapper>
              ))}
              path={V2Routes.GitRepo}
            />
            <Route
              component={withSearchParams((props: any) => (
                <CoreWrapper>
                  <WGApplicationsHelmRepository {...props} />
                </CoreWrapper>
              ))}
              path={V2Routes.HelmRepo}
            />
            <Route
              component={withSearchParams((props: any) => (
                <CoreWrapper>
                  <WGApplicationsBucket {...props} />
                </CoreWrapper>
              ))}
              path={V2Routes.Bucket}
            />
            <Route
              component={withSearchParams((props: any) => (
                <CoreWrapper>
                  <WGApplicationsHelmRelease {...props} />
                </CoreWrapper>
              ))}
              path={V2Routes.HelmRelease}
            />
            <Route
              component={withSearchParams((props: any) => (
                <CoreWrapper>
                  <WGApplicationsHelmChart {...props} />
                </CoreWrapper>
              ))}
              path={V2Routes.HelmChart}
            />
            <Route
              component={WGApplicationsFluxRuntime}
              exact
              path={V2Routes.FluxRuntime}
            />
            <Route exact component={Canaries} path="/applications/delivery" />

            <Route exact path={POLICIES} component={Policies} />
            <Route
              exact
              path="/policies/:id/:clusterName"
              component={PolicyDetails}
            />

            <Route
              exact
              path={GITLAB_OAUTH_CALLBACK}
              component={({ location }: any) => {
                const params = qs.parse(location.search);
                return (
                  <OAuthCallback
                    provider={'GitLab' as GitProvider}
                    code={params.code as string}
                  />
                );
              }}
            />
            <Route exact render={Page404} />
          </Switch>
        </main>
      </div>
    </Compose>
  );
};

const ResponsiveDrawer = () => {
  return (
    <AuthContextProvider>
      <EnterpriseClientProvider api={ClustersService}>
        <CoreClientContextProvider api={coreClient}>
          <Switch>
            <Route
              component={() => (
                <SignInWrapper>
                  <SignIn />
                </SignInWrapper>
              )}
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
        </CoreClientContextProvider>
      </EnterpriseClientProvider>
    </AuthContextProvider>
  );
};

export default ResponsiveDrawer;
