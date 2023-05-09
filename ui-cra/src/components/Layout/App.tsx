import { Drawer } from '@material-ui/core';
import CssBaseline from '@material-ui/core/CssBaseline';
import {
  AuthCheck,
  DetailModal,
  getParentNavRouteValue,
  Logo,
  SignIn,
  Layout,
  useNavigation,
} from '@weaveworks/weave-gitops';
import { Route, Switch } from 'react-router-dom';
import { ListConfigProvider, VersionProvider } from '../../contexts/ListConfig';
import AppRoutes from '../../routes';
import ErrorBoundary from '../ErrorBoundary';
import Navigation from './Navigation';
import { ToastContainer } from 'react-toastify';
import { useState } from 'react';
import { Routes } from '../../utils/nav';

const App = () => {
  const [collapsed, setCollapsed] = useState<boolean>(false);
  const { currentPage } = useNavigation();
  const value = getParentNavRouteValue(currentPage);

  return (
    <Layout
      logo={<Logo collapsed={collapsed} link={Routes.Clusters} />}
      nav={<Navigation />}
    >
      <ErrorBoundary>
        <ListConfigProvider>
          <VersionProvider>
            <AppRoutes />
            {/* <Drawer
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
              </Drawer> */}
          </VersionProvider>
        </ListConfigProvider>
      </ErrorBoundary>
      <ToastContainer
        position="top-center"
        autoClose={5000}
        newestOnTop={false}
      />
    </Layout>
  );
};

export default App;
