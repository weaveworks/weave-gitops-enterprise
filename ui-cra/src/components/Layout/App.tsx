import { Logo, Layout, Page as WGPage, Flex } from '@weaveworks/weave-gitops';
import { ListConfigProvider, VersionProvider } from '../../contexts/ListConfig';
import AppRoutes from '../../routes';
import ErrorBoundary from '../ErrorBoundary';
import Navigation from './Navigation';
import { ToastContainer } from 'react-toastify';
import { useState } from 'react';
import { Routes } from '../../utils/nav';
import styled from 'styled-components';
import { Breadcrumb } from '@weaveworks/weave-gitops/ui/components/Breadcrumbs';
import MemoizedHelpLinkWrapper from './HelpLinkWrapper';

export type PageProps = {
  className?: string;
  children?: any;
  loading?: boolean;
  path: Breadcrumb[];
};

const WGLayout = styled(Layout)`
  div[class*='Logo'] {
    height: 60px;
  }
`;

export const Page = ({ children, loading, className, path }: PageProps) => (
  <Flex column wide>
    <WGPage loading={loading} className={className} path={path}>
      {children}
    </WGPage>
    <MemoizedHelpLinkWrapper />
  </Flex>
);

const App = () => {
  const [collapsed, setCollapsed] = useState<boolean>(false);

  return (
    <ListConfigProvider>
      <VersionProvider>
        <WGLayout
          logo={<Logo collapsed={collapsed} link={Routes.Clusters} />}
          nav={<Navigation collapsed={collapsed} setCollapsed={setCollapsed} />}
        >
          <Flex column wide>
            <ErrorBoundary>
              <AppRoutes />
            </ErrorBoundary>
            <ToastContainer
              position="top-center"
              autoClose={5000}
              newestOnTop={false}
            />
            {/* <MemoizedHelpLinkWrapper /> */}
          </Flex>
        </WGLayout>
      </VersionProvider>
    </ListConfigProvider>
  );
};

export default App;
