import { Drawer } from '@material-ui/core';
import CssBaseline from '@material-ui/core/CssBaseline';
import { createStyles, makeStyles, Theme } from '@material-ui/core/styles';
import {
  AppContext,
  AuthCheck,
  DetailModal,
  SignIn,
} from '@weaveworks/weave-gitops';
import { useContext } from 'react';
import { Route, Routes } from 'react-router-dom';
import styled from 'styled-components';
import { ListConfigProvider, VersionProvider } from '../../contexts/ListConfig';
import AppRoutes from '../../routes';
import ErrorBoundary from '../ErrorBoundary';
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

const Layout = () => {
  const classes = useStyles();
  const { appState, setDetailModal } = useContext(AppContext);
  const detail = appState.detailModal;
  return (
    <Routes>
      <Route
        element={
          <SignInWrapper>
            <SignIn />
          </SignInWrapper>
        }
        path="/sign_in"
      />
      <Route
        path="*"
        element={
          <AuthCheck>
            <ListConfigProvider>
              <VersionProvider>
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
                      <DetailModal
                        className={detail.className}
                        object={detail.object}
                      />
                    )}
                  </Drawer>
                </div>
              </VersionProvider>
            </ListConfigProvider>
          </AuthCheck>
        }
      />
    </Routes>
  );
};

export default Layout;
