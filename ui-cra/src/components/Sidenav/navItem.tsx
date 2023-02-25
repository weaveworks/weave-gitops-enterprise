import { theme, useFeatureFlags, V2Routes } from '@weaveworks/weave-gitops';
import { NavLink, useLocation } from 'react-router-dom';
import styled from 'styled-components';
import { Routes } from '../../utils/nav';
import { ReactComponent as Applications } from '../assets/img/applications.svg';
import { ReactComponent as Clusters } from '../assets/img/clusters.svg';
import { ReactComponent as FluxIcon } from '../assets/img/flux-icon.svg';
import { ReactComponent as GitOpsRun } from '../assets/img/gitops-run-icon.svg';
import { ReactComponent as Policies } from '../assets/img/policies.svg';
import { ReactComponent as PolicyConfigs } from '../assets/img/policyConfigs.svg';
import { ReactComponent as SecretsIcon } from '../assets/img/secrets-Icon.svg';
import { ReactComponent as Templates } from '../assets/img/templates.svg';
import { ReactComponent as TerraformLogo } from '../assets/img/terraform-logo.svg';
import { ReactComponent as WorkspacesIcon } from '../assets/img/Workspace-Icon.svg';

interface SubNavItem {
  name: string;
  link: string;
  isVisible: boolean;
  relatedRoutes?: Array<string>;
}
interface NavigationItem {
  icon?: any;
  name: string;
  link: string;
  subItems?: Array<SubNavItem>;
  isVisible?: boolean;
  relatedRoutes?: Array<string>;
}

interface GroupNavItem {
  text: string;
  items: Array<NavigationItem>;
}

const { xxs, xs, small, medium } = theme.spacing;
const { neutral40, primary, neutral10, neutral30 } = theme.colors;
const NavWrapper = styled.div`
  display: flex;
  align-items: center;
  justify-content: start;
  flex-direction: column;
  margin-bottom: ${small};

  a.route-nav {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: start;
    padding-top: ${xs};
    padding-bottom: ${xs};
    padding-left: ${medium};
    padding-right: ${medium};
  }

  .parent-icon {
    width: 20px;
    height: 20px;
  }

  span.parent-route {
    margin-left: ${({ theme }) => theme.spacing.xs};
    letter-spacing: 1px;
  }

  a:not(a.nav-link-active):hover {
    background: ${neutral10};
  }
  .subroute-container {
    width: 100%;
  }

  .subroute-nav {
    padding: ${xs} ${xs} ${xs} calc(${medium} * 2 + ${xxs});
    color: ${neutral30};
    font-weight: 600;
  }
`;
export const NavItem = styled(NavLink).attrs({
  activeClassName: 'nav-link-active',
})`
  display: flex;
  font-size: ${12}px;
  box-sizing: border-box;
  color: ${neutral40};
  font-weight: bold;
  &.${props => props.activeClassName} {
    border-right: 3px solid ${primary};
    background: rgba(0, 179, 236, 0.1);
    color: ${primary};

    svg {
      fill: ${primary};

      &.gitops-run {
        stroke: ${primary};
      }
    }
  }
`;

const NavLinkItem = ({
  item,
  className,
}: {
  item: NavigationItem;
  className: string;
}) => {
  return (
    <NavItem
      exact={!!item.subItems ? true : false}
      to={item.link}
      className={`route-nav ${className}`}
    >
      <div className="parent-icon">{item.icon}</div>
      <span className="parent-route">{item.name}</span>
    </NavItem>
  );
};

const NavItems = () => {
  const { data: flagsRes } = useFeatureFlags();
  const location = useLocation();
  const groupItems: Array<GroupNavItem> = [
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
          isVisible: !!flagsRes.flags.WEAVE_GITOPS_FEATURE_PIPELINES,
        },
        {
          name: 'DELIVERY',
          link: Routes.Canaries,
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
          icon: <GitOpsRun />,
          isVisible: !!flagsRes.flags.WEAVE_GITOPS_FEATURE_RUN_UI,
        },
      ],
    },
  ];
  return (
    <>
      {groupItems.map(({ text, items }) => {
        return (
          <div key={text}>
            <h4>{text}</h4>
            {items.map(item => (
              <NavLinkItem
                key={item.name}
                item={item}
                className={
                  item.relatedRoutes?.some(link =>
                    location.pathname.includes(link),
                  )
                    ? 'nav-link-active'
                    : ''
                }
              />
            ))}
          </div>
        );
      })}
    </>
  );
};
export default NavItems;
