import CssBaseline from '@material-ui/core/CssBaseline';
import { createStyles, makeStyles, Theme } from '@material-ui/core/styles';
import { AuthCheck, SignIn } from '@weaveworks/weave-gitops';
import { Route, Switch } from 'react-router-dom';
import styled from 'styled-components';
import { ListConfigProvider, VersionProvider } from '../../contexts/ListConfig';
import NotificationsProvider from '../../contexts/Notifications/Provider';
import AppRoutes from '../../routes';
import ErrorBoundary from '../ErrorBoundary';
import Compose from '../ProvidersCompose';
import Navigation from './Navigation';

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      display: 'flex',
      width: '100vw',
    },
    toolbar: theme.mixins.toolbar,
    content: {
      flexGrow: 1,
      display: 'flex',
      width: '100%',
      overflow: 'hidden',
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
      </div>
    </Compose>
  );
};

const Layout = () => {
  return (
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
  );
};

export default Layout;
