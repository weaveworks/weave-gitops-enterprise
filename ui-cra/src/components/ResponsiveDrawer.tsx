import CssBaseline from '@material-ui/core/CssBaseline';
import Drawer from '@material-ui/core/Drawer';
import { createStyles, makeStyles, Theme } from '@material-ui/core/styles';
import {
  AppContext,
  AuthCheck,
  AuthContextProvider,
  coreClient,
  CoreClientContextProvider,
  DetailModal,
  LinkResolverProvider,
  Pendo,
  SignIn,
} from '@weaveworks/weave-gitops';
import React, { useContext } from 'react';
import { Route, Switch } from 'react-router-dom';
import styled from 'styled-components';
import { Terraform } from '../api/terraform/terraform.pb';
import { ClustersService } from '../cluster-services/cluster_services.pb';
import EnterpriseClientProvider from '../contexts/EnterpriseClient/Provider';
import { ListConfigProvider, VersionProvider } from '../contexts/ListConfig';
import NotificationsProvider from '../contexts/Notifications/Provider';
import { TerraformProvider } from '../contexts/Terraform';
import AppRoutes from '../routes';
import { resolver } from '../utils/link-resolver';
import ErrorBoundary from './ErrorBoundary';
import Compose from './ProvidersCompose';
import { Navigation } from './Sidenav/Navigation';

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      display: 'flex',
    },
    toolbar: theme.mixins.toolbar,
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
        <Navigation />
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
