import {
  Layout,
  Page as WGPage,
  Flex,
  useFeatureFlags,
  NavItem,
  IconType,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { ListConfigProvider, VersionProvider } from '../../contexts/ListConfig';
import AppRoutes from '../../routes';
import ErrorBoundary from '../ErrorBoundary';
import { ToastContainer } from 'react-toastify';
import { Routes } from '../../utils/nav';
import styled from 'styled-components';
import { Breadcrumb } from '@weaveworks/weave-gitops/ui/components/Breadcrumbs';
import MemoizedHelpLinkWrapper from './HelpLinkWrapper';
import { useMemo } from 'react';

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
  div[class*='Nav'] {
    height: 100vh;
  }
`;

function getNavItems(isFlagEnabled: (flag: string) => boolean): NavItem[] {
  return [
    {
      label: 'Platform',
    },
    {
      label: 'Clusters',
      link: { value: Routes.Clusters },
      icon: IconType.ClustersIcon,
    },
    {
      label: 'Templates',
      link: { value: Routes.Templates },
      icon: IconType.TemplatesIcon,
    },
    {
      label: 'GitOpsSets',
      link: { value: Routes.GitOpsSets },
      icon: IconType.GitOpsSetsIcon,
    },
    {
      label: 'Terraform',
      link: { value: Routes.TerraformObjects },
      icon: IconType.TerraformIcon,
      disabled: !isFlagEnabled('WEAVE_GITOPS_FEATURE_TERRAFORM_UI'),
    },
    {
      label: 'Secrets',
      link: { value: Routes.Secrets },
      icon: IconType.SecretsIcon,
    },
    {
      label: 'Delivery',
    },
    {
      label: 'Applications',
      link: { value: V2Routes.Automations },
      icon: IconType.ApplicationsIcon,
    },
    {
      label: 'Sources',
      link: { value: V2Routes.Sources },
      icon: IconType.SourcesIcon,
    },
    {
      label: 'Image Automation',
      link: { value: Routes.ImageAutomation },
      icon: IconType.ImageAutomationIcon,
    },
    {
      label: 'Pipelines',
      link: { value: Routes.Pipelines },
      icon: IconType.PipelinesIcon,
      disabled: !isFlagEnabled('WEAVE_GITOPS_FEATURE_PIPELINES'),
    },
    {
      label: 'Delivery ',
      link: { value: Routes.Canaries },
      icon: IconType.DeliveryIcon,
      disabled: process.env.REACT_APP_DISABLE_PROGRESSIVE_DELIVERY === 'true',
    },
    {
      label: 'Flux Runtime',
      link: { value: V2Routes.FluxRuntime },
      icon: IconType.FluxIcon,
    },
    {
      label: 'Explorer',
      link: { value: Routes.Explorer },
      icon: IconType.ExploreIcon,
      disabled: !isFlagEnabled('WEAVE_GITOPS_FEATURE_EXPLORER'),
    },
    {
      label: 'Guardrails',
    },
    {
      label: 'Workspaces',
      link: { value: Routes.Workspaces },
      icon: IconType.WorkspacesIcon,
    },
    {
      label: 'Policies',
      link: { value: Routes.Policies },
      icon: IconType.PoliciesIcon,
    },
    {
      label: 'Policy Configs',
      link: { value: Routes.PolicyConfigs },
      icon: IconType.PolicyConfigsIcon,
    },
    {
      label: 'Developer Ex.',
    },
    {
      label: 'GitOps Run',
      link: { value: Routes.GitOpsRun },
      icon: IconType.GitOpsRunIcon,
    },
    {
      label: 'Notifications',
      link: { value: Routes.Notifications },
      icon: IconType.NotificationsIcon,
    },
  ];
}

export const Page = ({ children, loading, className, path }: PageProps) => (
  <Flex column wide>
    <WGPage loading={loading} className={className} path={path}>
      {children}
    </WGPage>
    <MemoizedHelpLinkWrapper />
  </Flex>
);

const App = () => {
  const { isFlagEnabled } = useFeatureFlags();
  const navItems = useMemo(() => getNavItems(isFlagEnabled), [isFlagEnabled]);

  return (
    <ListConfigProvider>
      <VersionProvider>
        <WGLayout logoLink={Routes.Clusters} navItems={navItems}>
          <Flex column wide>
            <ErrorBoundary>
              <AppRoutes />
            </ErrorBoundary>
            <ToastContainer
              position="top-center"
              autoClose={5000}
              newestOnTop={false}
            />
          </Flex>
        </WGLayout>
      </VersionProvider>
    </ListConfigProvider>
  );
};

export default App;
