import Box from '@material-ui/core/Box';
import { makeStyles } from '@material-ui/core/styles';
import {
  Link,
  theme,
  useFeatureFlags,
  V2Routes,
} from '@weaveworks/weave-gitops';
import React, { FC } from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import styled from 'styled-components';
import WeaveGitOps from '../assets/img/weave-logo.svg';
import { useListConfigContext } from '../contexts/ListConfig';
import { Routes } from '../utils/nav';
import SVGIcon, { IconType } from './SVGIcon';

const { xxs, xs, small, medium } = theme.spacing;
const { neutral10, neutral30, neutral40, primary } = theme.colors;

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

const useStyles = makeStyles({
  root: {
    paddingTop: medium,
    alignItems: 'center',
    height: 'calc(100vh - 84px)',
    borderTopRightRadius: '10px',
  },
  logo: {
    padding: `calc(${medium} - ${xxs})`,
    paddingTop: `${medium}`,
    paddingBottom: `17px`,
  },
});

const NavItems = () => {
  const { data: flagsRes } = useFeatureFlags();
  const location = useLocation();
  const navItems: Array<NavigationItem> = [
    {
      name: 'CLUSTERS',
      link: Routes.Clusters,
      icon: <SVGIcon icon={IconType.Clusters} />,
      subItems: [
        {
          name: 'VIOLATION LOG',
          link: Routes.PolicyViolations,
          isVisible: true,
        },
      ],
    },
    {
      name: 'APPLICATIONS',
      link: V2Routes.Automations,
      icon: <SVGIcon icon={IconType.Applications} />,
      subItems: [
        {
          name: 'SOURCES',
          link: V2Routes.Sources,
          isVisible: true,
          relatedRoutes: [
            V2Routes.GitRepo,
            V2Routes.HelmRepo,
            V2Routes.OCIRepository,
            V2Routes.HelmChart,
          ],
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
            import.meta.env.REACT_APP_DISABLE_PROGRESSIVE_DELIVERY !== 'true',
          relatedRoutes: [Routes.CanaryDetails],
        },
      ],
      relatedRoutes: [V2Routes.Kustomization, V2Routes.HelmRelease],
    },

    {
      name: 'GITOPS SETS',
      link: Routes.GitOpsSets,
      icon: <SVGIcon icon={IconType.GitOpsSetsIcon} />,
    },
    {
      name: 'GITOPS RUN',
      link: Routes.GitOpsRun,
      icon: <SVGIcon icon={IconType.GitOpsRun} className="gitops-run" />,
      isVisible: !!flagsRes.flags.WEAVE_GITOPS_FEATURE_RUN_UI,
    },
    {
      name: 'TEMPLATES',
      link: Routes.Templates,
      icon: <SVGIcon icon={IconType.Templates} />,
    },
    {
      name: 'TERRAFORM',
      link: Routes.TerraformObjects,
      icon: <SVGIcon icon={IconType.TerraformLogo} />,
      isVisible: !!flagsRes.flags.WEAVE_GITOPS_FEATURE_TERRAFORM_UI,
    },
    {
      name: 'WORKSPACES',
      link: Routes.Workspaces,
      icon: <SVGIcon icon={IconType.Workspaces} />,
    },
    {
      name: 'FLUX RUNTIME',
      link: V2Routes.FluxRuntime,
      icon: <SVGIcon icon={IconType.Flux} />,
    },
    {
      name: 'SECRETS',
      link: Routes.Secrets,
      icon: <SVGIcon icon={IconType.SecretsIcon} />,
    },
    {
      name: 'POLICIES',
      link: Routes.Policies,
      icon: <SVGIcon icon={IconType.Policies} />,
    },
    {
      name: 'POLICY CONFIGS',
      link: Routes.PolicyConfigs,
      icon: <SVGIcon icon={IconType.PolicyConfigs} />,
    },
  ];
  return (
    <>
      {navItems.map(item => {
        return item.isVisible !== false ? (
          <NavWrapper key={item.name}>
            <NavItem
              exact={!!item.subItems ? true : false}
              to={item.link}
              className={`route-nav ${
                item.relatedRoutes?.some(link =>
                  location.pathname.includes(link),
                )
                  ? 'nav-link-active'
                  : ''
              }`}
            >
              <div className="parent-icon">{item.icon}</div>
              <span className="parent-route">{item.name}</span>
            </NavItem>

            {item.subItems && (
              <div className="subroute-container">
                {item.subItems?.map(subItem => {
                  return (
                    subItem.isVisible && (
                      <NavItem
                        to={subItem.link}
                        key={subItem.name}
                        className={`subroute-nav ${
                          subItem.relatedRoutes?.some(link =>
                            location.pathname.includes(link),
                          )
                            ? 'nav-link-active'
                            : ''
                        }`}
                      >
                        {subItem.name}
                      </NavItem>
                    )
                  );
                })}
              </div>
            )}
          </NavWrapper>
        ) : null;
      })}
    </>
  );
};

const MemoizedNavItems = React.memo(NavItems);

export const Navigation: FC = () => {
  const classes = useStyles();
  const listConfigContext = useListConfigContext();
  const uiConfig = listConfigContext?.uiConfig || '';

  return (
    <>
      <div title="Home" className={classes.logo}>
        <Link to={Routes.Clusters}>
          <img src={uiConfig?.logoURL || WeaveGitOps} alt="Home" />
        </Link>
      </div>
      <Box className={`${classes.root} nav-items`} bgcolor={theme.colors.white}>
        <MemoizedNavItems />
      </Box>
    </>
  );
};
