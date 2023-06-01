import { Logo, Layout } from '@weaveworks/weave-gitops';
import { ListConfigProvider, VersionProvider } from '../../contexts/ListConfig';
import AppRoutes from '../../routes';
import ErrorBoundary from '../ErrorBoundary';
import Navigation from './Navigation';
import { ToastContainer } from 'react-toastify';
import { useState } from 'react';
import { Routes } from '../../utils/nav';
import styled from 'styled-components';

const WGLayout = styled(Layout)`
  div[class*='Logo'] {
    height: 60px;
  }
`;

const App = () => {
  const [collapsed, setCollapsed] = useState<boolean>(false);

  return (
    <ListConfigProvider>
      <VersionProvider>
        <WGLayout
          logo={<Logo collapsed={collapsed} link={Routes.Clusters} />}
          nav={<Navigation collapsed={collapsed} setCollapsed={setCollapsed} />}
        >
          <ErrorBoundary>
            <AppRoutes />
          </ErrorBoundary>
          <ToastContainer
            position="top-center"
            autoClose={5000}
            newestOnTop={false}
          />
        </WGLayout>
      </VersionProvider>
    </ListConfigProvider>
  );
};

export default App;
