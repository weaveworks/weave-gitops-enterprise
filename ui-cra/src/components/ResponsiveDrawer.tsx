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
  AppContext,
  AuthCheck,
  AuthContextProvider,
  coreClient,
  CoreClientContextProvider,
  LinkResolverProvider,
  Pendo,
  SignIn,
  theme as weaveTheme,
  DetailModal,
} from '@weaveworks/weave-gitops';
import React, { useContext } from 'react';
import { Route, Switch } from 'react-router-dom';
import styled from 'styled-components';
import { Terraform } from '../api/terraform/terraform.pb';
import { ReactComponent as MenuIcon } from '../assets/img/menu-burger.svg';
import { ClustersService } from '../cluster-services/cluster_services.pb';
import EnterpriseClientProvider from '../contexts/EnterpriseClient/Provider';
import { TerraformProvider } from '../contexts/Terraform';
import NotificationsProvider from '../contexts/Notifications/Provider';
import AppRoutes from '../routes';
import { Navigation } from './Navigation';
import Compose from './ProvidersCompose';
import { resolver } from '../utils/link-resolver';
import { ListConfigProvider, VersionProvider } from '../contexts/ListConfig';
import ErrorBoundary from './ErrorBoundary';

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

const App = () => {
  const classes = useStyles();
  const theme = useTheme();
  const [mobileOpen, setMobileOpen] = React.useState(false);

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };
  const { appState, setDetailModal } = useContext(AppContext);

  const detail = appState.detailModal;

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
            <AppRoutes />
          </ErrorBoundary>
        </main>
        <Drawer
          anchor="right"
          open={detail ? true : false}
          onClose={() => setDetailModal(null)}
          ModalProps={{ keepMounted: false }}
        >
          {detail && (
            <DetailModal className={detail.className} object={detail.object} />
          )}
        </Drawer>
      </div>
    </Compose>
  );
};

const ResponsiveDrawer = () => {
  return (
    <AuthContextProvider>
      <EnterpriseClientProvider api={ClustersService}>
        <CoreClientContextProvider api={coreClient}>
          <TerraformProvider api={Terraform}>
            <LinkResolverProvider resolver={resolver}>
              <Pendo
                defaultTelemetryFlag="true"
                //@ts-ignore
                tier="enterprise"
                version={process.env.REACT_APP_VERSION}
              />
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
                    <ListConfigProvider>
                      <VersionProvider>
                        <App />
                      </VersionProvider>
                    </ListConfigProvider>
                  </AuthCheck>
                </Route>
              </Switch>
            </LinkResolverProvider>
          </TerraformProvider>
        </CoreClientContextProvider>
      </EnterpriseClientProvider>
    </AuthContextProvider>
  );
};

export default ResponsiveDrawer;
