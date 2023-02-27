import { V2Routes } from '@weaveworks/weave-gitops';
import { ReactComponent as Applications } from '../../assets/img/menu-icon/applications.svg';
import { ReactComponent as Clusters } from '../../assets/img/menu-icon/cluster.svg';
import { ReactComponent as FluxIcon } from '../../assets/img/menu-icon/fluxruntime.svg';
import { ReactComponent as GitOpsRun } from '../../assets/img/menu-icon/gitopsrun.svg';
import { ReactComponent as Policies } from '../../assets/img/menu-icon/policies.svg';
import { ReactComponent as PolicyConfigs } from '../../assets/img/menu-icon/policyConfigs.svg';
import { Routes } from '../../utils/nav';

import { ReactComponent as DeliveryIcon } from '../../assets/img/menu-icon/delivery.svg';
import { ReactComponent as ImageAutomationIcon } from '../../assets/img/menu-icon/imageautomation.svg';
import { ReactComponent as NotificationIcon } from '../../assets/img/menu-icon/notification.svg';
import { ReactComponent as PipeineIcon } from '../../assets/img/menu-icon/pipelines.svg';
import { ReactComponent as SecretsIcon } from '../../assets/img/menu-icon/secrets.svg';
import { ReactComponent as Templates } from '../../assets/img/menu-icon/templates.svg';
import { ReactComponent as TerraformLogo } from '../../assets/img/menu-icon/terraform.svg';
import { ReactComponent as WorkspacesIcon } from '../../assets/img/menu-icon/workspaces.svg';

export interface NavigationItem {
  icon?: any;
  name: string;
  link: string;
  isVisible?: boolean;
  relatedRoutes?: Array<string>;
}

export interface GroupNavItem {
  text: string;
  items: Array<NavigationItem>;
}
export function getNavItems(flagsRes: {
  flags: {
    [key: string]: string;
  };
}): GroupNavItem[] {
  return [
    {
      text: 'PLATFORM',
      items: [
        {
          name: 'CLUSTERS',
          link: Routes.Clusters,
          icon: <Clusters />,
        },
        {
          name: 'TEMPLATES',
          link: Routes.Templates,
          icon: <Templates />,
        },
        {
          name: 'TERRAFORM',
          link: Routes.TerraformObjects,
          icon: <TerraformLogo />,
          isVisible: !!flagsRes.flags.WEAVE_GITOPS_FEATURE_TERRAFORM_UI,
        },
        {
          name: 'SECRETS',
          link: Routes.Secrets,
          icon: <SecretsIcon />,
        },
      ],
    },
    {
      text: 'DELIVERY',
      items: [
        {
          name: 'APPLICATIONS',
          link: V2Routes.Automations,
          icon: <Applications />,
          relatedRoutes: [V2Routes.Kustomization, V2Routes.HelmRelease],
        },
        {
          name: 'IMAGE AUTOMATION',
          link: Routes.ImageAutomation,
          icon: <ImageAutomationIcon />,
          isVisible: true,
          relatedRoutes: [
            V2Routes.ImageAutomationRepositoryDetails,
            V2Routes.ImagePolicyDetails,
            V2Routes.ImageAutomationUpdatesDetails,
          ],
        },
        {
          name: 'PIPELINES',
          link: Routes.Pipelines,
          icon: <PipeineIcon />,
          isVisible: !!flagsRes.flags.WEAVE_GITOPS_FEATURE_PIPELINES,
        },
        {
          name: 'DELIVERY',
          link: Routes.Canaries,
          icon: <DeliveryIcon />,
          isVisible:
            process.env.REACT_APP_DISABLE_PROGRESSIVE_DELIVERY !== 'true',
          relatedRoutes: [Routes.CanaryDetails],
        },
        {
          name: 'FLUX RUNTIME',
          link: V2Routes.FluxRuntime,
          icon: <FluxIcon />,
        },
      ],
    },
    {
      text: 'GUARDRAILS',
      items: [
        {
          name: 'WORKSPACES',
          link: Routes.Workspaces,
          icon: <WorkspacesIcon />,
        },
        {
          name: 'POLICIES',
          link: Routes.Policies,
          icon: <Policies />,
        },
        {
          name: 'POLICY CONFIGS',
          link: Routes.PolicyConfigs,
          icon: <PolicyConfigs />,
        },
      ],
    },
    {
      text: 'DEVELOPER EX.',
      items: [
        {
          name: 'GITOPS RUN',
          link: Routes.GitOpsRun,
          icon: <GitOpsRun className="gitops-run" />,
          isVisible: !!flagsRes.flags.WEAVE_GITOPS_FEATURE_RUN_UI,
        },
        {
          name: 'NOTIFICATIONS',
          link: Routes.Notifications,
          icon: <NotificationIcon />,
          isVisible: !!flagsRes.flags.WEAVE_GITOPS_FEATURE_RUN_UI,
        },
      ],
    },
  ];
}
