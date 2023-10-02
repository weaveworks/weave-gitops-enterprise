import {
  getParentNavRouteValue,
  IconType,
  Nav,
  NavItem,
  useFeatureFlags,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import { Dispatch, FC, SetStateAction, useMemo } from 'react';
import styled from 'styled-components';
import { Routes } from '../../utils/nav';
import { useLocation } from 'react-router-dom';

function getParentNavRouteValueExtended(
  route: string,
): V2Routes | Routes | boolean | PageRoute {
  if (route === V2Routes.PolicyViolationDetails) {
    if (!window.location.search.includes('kind')) return Routes.Clusters;
    if (window.location.search.includes('kind=Policy')) {
      return Routes.Policies;
    }
    return V2Routes.Automations;
  }

  if (Object.values(V2Routes).includes(route as V2Routes)) {
    return getParentNavRouteValue(route);
  }

  switch (route) {
    //Clusters
    case Routes.Clusters:
    case Routes.ClusterDashboard:
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
    //GitOps Sets
    case Routes.GitOpsSets:
    case Routes.GitOpsSetDetail:
      return Routes.GitOpsSets;

    case Routes.Explorer:
    case Routes.ExplorerAccessRules:
      return Routes.Explorer;

    default:
      return false;
  }
}

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
      label: 'Progressive Delivery ',
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

const Navigation: FC<{
  collapsed: boolean;
  setCollapsed: Dispatch<SetStateAction<boolean>>;
}> = ({ collapsed, setCollapsed }) => {
  const { isFlagEnabled } = useFeatureFlags();
  const navItems = useMemo(() => getNavItems(isFlagEnabled), [isFlagEnabled]);
  const currentPage = useLocation();
  const routeValue = getParentNavRouteValueExtended(
    '/' + currentPage.pathname.split('/')[1],
  );

  return (
    <Nav
      className="test-id-navigation"
      collapsed={collapsed}
      setCollapsed={setCollapsed}
      navItems={navItems}
      currentPage={routeValue}
    />
  );
};

export default styled(Navigation)`
  height: 100%;
`;
