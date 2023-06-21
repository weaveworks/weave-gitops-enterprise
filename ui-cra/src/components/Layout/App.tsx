import { Layout, Logo, Page as WGPage } from '@weaveworks/weave-gitops';
import { Breadcrumb } from '@weaveworks/weave-gitops/ui/components/Breadcrumbs';
import { useState } from 'react';
import { ToastContainer } from 'react-toastify';
import { ListConfigProvider, VersionProvider } from '../../contexts/ListConfig';
import AppRoutes from '../../routes';
import { Routes } from '../../utils/nav';
import ErrorBoundary from '../ErrorBoundary';
import Navigation from './Navigation';

export type PageProps = {
  className?: string;
  children?: any;
  loading?: boolean;
  path: Breadcrumb[];
};

export const Page = ({ children, loading, className, path }: PageProps) => (
  <WGPage loading={loading} className={className} path={path}>
    {children}
  </WGPage>
);

const App = () => {
  const [collapsed, setCollapsed] = useState<boolean>(false);
  const logo = <Logo collapsed={collapsed} link={Routes.Clusters} />;
  const nav = <Navigation collapsed={collapsed} setCollapsed={setCollapsed} />;

  return (
    <ListConfigProvider>
      <VersionProvider>
        <Layout logo={logo} nav={nav}>
          <ErrorBoundary>
            <AppRoutes />
          </ErrorBoundary>
          <ToastContainer
            position="top-center"
            autoClose={5000}
            newestOnTop={false}
          />
        </Layout>
      </VersionProvider>
    </ListConfigProvider>
  );
};

export default App;
