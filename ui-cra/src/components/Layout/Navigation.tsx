import {
  getParentNavRouteValue,
  IconType,
  Logo,
  Nav,
  NavItem,
  useFeatureFlags,
} from '@weaveworks/weave-gitops';
import { FC, useMemo, useState } from 'react';

import { V2Routes } from '@weaveworks/weave-gitops';

import { Routes } from '../../utils/nav';

import { FeatureFlags } from '@weaveworks/weave-gitops/ui/hooks/featureflags';
import { useLocation } from 'react-router-dom';
import styled from 'styled-components';

function getParentNavRouteValueExtended(
  route: string,
): V2Routes | Routes | boolean {
  if (Object.values(V2Routes).includes(route as V2Routes)) {
    console.log(getParentNavRouteValue(route));
    return getParentNavRouteValue(route);
  }
  switch (route) {
    //Clusters
    case Routes.ClusterDashboard:
    case Routes.PolicyViolations:
    case Routes.PolicyViolationDetails:
    case Routes.DeleteCluster:
    case Routes.EditResource:
      return Routes.Clusters;
    //Templates
    case Routes.Templates:
    case Routes.AddCluster:
    case Routes.AddApplication:
      return Routes.Templates;
    //Secrets
    case Routes.Secrets:
    case Routes.SecretDetails:
    case Routes.CreateSecret:
      return Routes.Secrets;
    //Terraform
    case Routes.TerraformObjects:
    case Routes.TerraformDetail:
      return Routes.TerraformObjects;
    //Pipelines
    case Routes.Pipelines:
    case Routes.PipelineDetails:
      return Routes.Pipelines;
    //Canaries
    case Routes.Canaries:
    case Routes.CanaryDetails:
      return Routes.Canaries;
    //Workspaces
    case Routes.Workspaces:
    case Routes.WorkspaceDetails:
      return Routes.Workspaces;
    //Policy
    case Routes.Policies:
    case Routes.PolicyDetails:
    case Routes.GitlabOauthCallback:
    case Routes.BitBucketOauthCallback:
      return Routes.Policies;
    //Policy Configs
    case Routes.PolicyConfigs:
      return Routes.PolicyConfigs;
    //GitOps Run
    case Routes.GitOpsRun:
    case Routes.GitOpsRunDetail:
      return Routes.GitOpsRun;
    default:
      return false;
  }
}

function getNavItems(flagsRes: FeatureFlags): NavItem[] {
  return [
    {
      label: 'Platform',
    },

    {
      label: 'CLUSTERS',
      link: { value: Routes.Clusters },
      icon: IconType.DnsOutlined,
    },
    {
      label: 'TEMPLATES',
      link: { value: Routes.Templates },
      icon: IconType.DashboardOutlined,
    },
    {
      label: 'TERRAFORM',
      link: { value: Routes.TerraformObjects },
      icon: IconType.TerraformIcon,
      disabled: !!flagsRes.WEAVE_GITOPS_FEATURE_TERRAFORM_UI,
    },
    {
      label: 'SECRETS',
      link: { value: Routes.Secrets },
      icon: IconType.VpnKeyOutlined,
    },

    {
      label: 'Delivery',
    },

    {
      label: 'APPLICATIONS',
      link: { value: V2Routes.Automations },
      icon: IconType.ApplicationsIcon,
    },
    {
      label: 'SOURCES',
      link: { value: V2Routes.Sources },
      icon: IconType.SourcesIcon,
    },
    {
      label: 'IMAGE AUTOMATION',
      link: { value: Routes.ImageAutomation },
      icon: IconType.ImageAutomationIcon,
    },
    {
      label: 'PIPELINES',
      link: { value: Routes.Pipelines },
      icon: IconType.PipelinesIcon,
      disabled: !!flagsRes.WEAVE_GITOPS_FEATURE_PIPELINES,
    },
    {
      label: 'DELIVERY',
      link: { value: Routes.Canaries },
      icon: IconType.DeliveryIcon,
      disabled: process.env.REACT_APP_DISABLE_PROGRESSIVE_DELIVERY !== 'true',
    },
    {
      label: 'FLUX RUNTIME',
      link: { value: V2Routes.FluxRuntime },
      icon: IconType.FluxIcon,
    },

    {
      label: 'GUARDRAILS',
    },

    {
      label: 'WORKSPACES',
      link: { value: Routes.Workspaces },
      icon: IconType.TabOutlined,
    },
    {
      label: 'POLICIES',
      link: { value: Routes.Policies },
      icon: IconType.PolicyOutlined,
    },
    {
      label: 'POLICY CONFIGS',
      link: { value: Routes.PolicyConfigs },
      icon: IconType.VerifiedUserOutlined,
    },

    {
      label: 'DEVELOPER EX.',
    },

    {
      label: 'GITOPS RUN',
      link: { value: Routes.GitOpsRun },
      icon: IconType.GitOpsRunIcon,
    },
    {
      label: 'NOTIFICATIONS',
      link: { value: Routes.Notifications },
      icon: IconType.NotificationsBell,
    },
  ];
}

const LogoHeight = styled.div`
  display: flex;
  align-items: center;
  height: 60px;
`;

const NavHeight = styled.div`
  height: calc(100% - 60px);
`;

const NavContainer = styled.div`
  height: 100vh;
`;

const Navigation: FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const { data } = useFeatureFlags();
  const navItems = useMemo(() => getNavItems(data.flags), [data]);
  const currentPage = useLocation();
  const routeValue = getParentNavRouteValueExtended(
    '/' + currentPage.pathname.split('/')[1],
  );
  return (
    <NavContainer>
      <LogoHeight>
        <Logo link={Routes.Clusters} collapsed={collapsed} />
      </LogoHeight>
      <NavHeight>
        <Nav
          collapsed={collapsed}
          setCollapsed={setCollapsed}
          navItems={navItems}
          currentPage={routeValue}
        />
      </NavHeight>
    </NavContainer>
  );
};

export default styled(Navigation)`
  height: 100%;
`;
