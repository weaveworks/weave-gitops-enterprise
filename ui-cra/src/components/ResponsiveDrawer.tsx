import React from 'react';
import { Switch, Route, Redirect } from 'react-router-dom';
import MCCP from './Clusters';
import TemplatesDashboard from './Templates';
import { Navigation } from './Navigation';
import CssBaseline from '@material-ui/core/CssBaseline';
import Drawer from '@material-ui/core/Drawer';
import Hidden from '@material-ui/core/Hidden';
import IconButton from '@material-ui/core/IconButton';
import { ReactComponent as MenuIcon } from '../assets/img/menu-burger.svg';
import {
  makeStyles,
  useTheme,
  Theme,
  createStyles,
} from '@material-ui/core/styles';
import {
  AuthContextProvider,
  AuthCheck,
  coreClient,
  CoreClientContextProvider,
  OAuthCallback,
  SignIn,
  V2Routes,
} from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import NotificationsProvider from '../contexts/Notifications/Provider';
import Compose from './ProvidersCompose';
import Box from '@material-ui/core/Box';
import { PageTemplate } from './Layout/PageTemplate';
import { SectionHeader } from './Layout/SectionHeader';
import { ContentWrapper } from './Layout/ContentWrapper';
import Lottie from 'react-lottie-player';
import error404 from '../assets/img/error404.json';
import AddClusterWithCredentials from './Clusters/Create';
import EditCluster from './Clusters/Edit';
import WGApplicationsDashboard from './Applications';
import WGApplicationsSources from './Applications/Sources';
import WGApplicationsKustomization from './Applications/Kustomization';
import WGApplicationsGitRepository from './Applications/GitRepository';
import WGApplicationsHelmRepository from './Applications/HelmRepository';
import WGApplicationsBucket from './Applications/Bucket';
import WGApplicationsHelmRelease from './Applications/HelmRelease';
import WGApplicationsHelmChart from './Applications/HelmChart';
import WGApplicationsFluxRuntime from './Applications/FluxRuntime';
import qs from 'query-string';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import Policies from './Policies';
import PolicyDetails from './Policies/PolicyDetails';
import PoliciesViolations from './PolicyViolations';
import PolicyViolationDetails from './PolicyViolations/ViolationDetails';
import { ClustersService } from '../cluster-services/cluster_services.pb';
import EnterpriseClientProvider from '../contexts/EnterpriseClient/Provider';
import ProgressiveDelivery from './ProgressiveDelivery';
import ClusterDashboard from './Clusters/ClusterDashboard';
import ErrorBoundary from './ErrorBoundary';
import CanaryDetails from './ProgressiveDelivery/CanaryDetails';
import AddApplication from './Applications/Add';
import Pipelines from './Pipelines';
import WGApplicationsOCIRepository from './Applications/OCIRepository';
import PipelineDetails from './Pipelines/PipelineDetails';

const GITLAB_OAUTH_CALLBACK = '/oauth/gitlab';
const POLICIES = '/policies';
const CANARIES = '/applications/delivery';
const PIPELINES = '/applications/pipelines';
export const EDIT_CLUSTER = '/clusters/:clusterName/edit';

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
      background: 'none',
    },
    content: {
      flexGrow: 1,
    },
  }),
);

const SignInWrapper = styled.div`
  height: calc(100vh - 90px);
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
    <SectionHeader path={[{ label: 'Error' }]} />
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
    <Compose components={[NotificationsProvider]}>
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
              <div>
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
              <Navigation />
            </Drawer>
          </Hidden>
        </nav>
        <main className={classes.content}>
          <ErrorBoundary>
            <Switch>
              <Route exact path="/">
                <Redirect to="/clusters" />
              </Route>
              <Route component={MCCP} exact path={'/clusters'} />
              <Route component={MCCP} exact path="/clusters/delete" />
              <Route
                component={withSearchParams((props: any) => (
                  <ClusterDashboard {...props} />
                ))}
                path="/cluster"
              />
              <Route component={EditCluster} exact path={EDIT_CLUSTER} />
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
                component={withSearchParams(PolicyViolationDetails)}
                exact
                path="/clusters/violations/details"
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
                component={AddApplication}
                exact
                path="/applications/create"
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
                component={withSearchParams((props: any) => (
                  <CoreWrapper>
                    <WGApplicationsOCIRepository {...props} />
                  </CoreWrapper>
                ))}
                path={V2Routes.OCIRepository}
              />
              <Route
                component={() => (
                  <CoreWrapper>
                    <WGApplicationsFluxRuntime />
                  </CoreWrapper>
                )}
                path={V2Routes.FluxRuntime}
              />
              <Route exact path={CANARIES} component={ProgressiveDelivery} />
              <Route exact path={PIPELINES} component={Pipelines} />
              <Route
                exact
                path="/applications/pipelines/details"
                component={withSearchParams(PipelineDetails)}
              />
              <Route
                path="/applications/delivery/:id"
                component={withSearchParams(CanaryDetails)}
              />
              <Route exact path={POLICIES} component={Policies} />
              <Route
                exact
                path="/policies/details"
                component={withSearchParams(PolicyDetails)}
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
          </ErrorBoundary>
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
