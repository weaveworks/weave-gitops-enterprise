import { Layout, Page as WGPage, Flex, Logo } from '@weaveworks/weave-gitops';
import { ListConfigProvider, VersionProvider } from '../../contexts/ListConfig';
import AppRoutes from '../../routes';
import ErrorBoundary from '../ErrorBoundary';
import { ToastContainer } from 'react-toastify';
import { Routes } from '../../utils/nav';
import styled from 'styled-components';
import { Breadcrumb } from '@weaveworks/weave-gitops/ui/components/Breadcrumbs';
import { useState } from 'react';
import Navigation from './Navigation';

export type PageProps = {
  className?: string;
  children?: any;
  loading?: boolean;
  path: Breadcrumb[];
};

const WGLayout = styled(Layout)`
  div[class*='Nav__NavContainer'] {
    height: calc(100vh - 60px);
  }

  //TO DO: Remove once dark mode is implemented
  div[class*='UserSettings'] {
    span[class*='MuiSwitch-root'] {
      visibility: hidden;
    }
  }
`;

export const Page = ({ children, loading, className, path }: PageProps) => (
  <Flex column wide>
    <WGPage loading={loading} className={className} path={path}>
      {children}
    </WGPage>
  </Flex>
);

const App = () => {
  const [collapsed, setCollapsed] = useState<boolean>(false);
  const logo = <Logo collapsed={collapsed} link={Routes.Clusters} />;
  const nav = <Navigation collapsed={collapsed} setCollapsed={setCollapsed} />;

  return (
    <ListConfigProvider>
      <VersionProvider>
        <WGLayout logo={logo} nav={nav}>
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
