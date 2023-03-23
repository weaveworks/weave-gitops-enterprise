import {
  Flex,
  getParentNavRouteValue,
  IconType,
  NavItem,
  V2Routes
} from '@weaveworks/weave-gitops';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import styled from 'styled-components';
import { Routes } from '../../utils/nav';

function getParentNavRouteValueExtended(
  route: string,
): V2Routes | Routes | boolean | PageRoute {
  if (Object.values(V2Routes).includes(route as V2Routes)) {
    return getParentNavRouteValue(route);
  }
  switch (route) {
    //Clusters
    case Routes.Clusters:
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

function getNavItems(isFlagEnabled: (flag: string) => boolean): NavItem[] {
  return [
    {
      label: 'Platform',
    },
    {
      label: 'Clusters',
      link: { value: Routes.Clusters },
      icon: IconType.DnsOutlined,
    },
    {
      label: 'Templates',
      link: { value: Routes.Templates },
      icon: IconType.DashboardOutlined,
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
      icon: IconType.VpnKeyOutlined,
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
      label: 'Guardrails',
    },
    {
      label: 'Workspaces',
      link: { value: Routes.Workspaces },
      icon: IconType.TabOutlined,
    },
    {
      label: 'Policies',
      link: { value: Routes.Policies },
      icon: IconType.PolicyOutlined,
    },
    {
      label: 'Policy Configs',
      link: { value: Routes.PolicyConfigs },
      icon: IconType.VerifiedUserOutlined,
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
      icon: IconType.NotificationsBell,
    },
  ];
}

const LogoHeight = styled(Flex)`
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

  const { isFlagEnabled } = useFeatureFlags();
  const navItems = useMemo(() => getNavItems(isFlagEnabled), [isFlagEnabled]);

  const currentPage = useLocation();
  const routeValue = getParentNavRouteValueExtended(
    '/' + currentPage.pathname.split('/')[1],
  );

  return (
    <NavContainer>
      <LogoHeight align>
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
