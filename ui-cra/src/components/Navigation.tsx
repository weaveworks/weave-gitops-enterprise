import Box from '@material-ui/core/Box';
import { makeStyles } from '@material-ui/core/styles';
import { theme, useFeatureFlags, V2Routes } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { NavLink } from 'react-router-dom';
import styled from 'styled-components';
import { ReactComponent as Applications } from '../assets/img/applications.svg';
import { ReactComponent as Clusters } from '../assets/img/clusters.svg';
import { ReactComponent as FluxIcon } from '../assets/img/flux-icon.svg';
import { ReactComponent as Policies } from '../assets/img/policies.svg';
import { ReactComponent as Templates } from '../assets/img/templates.svg';
import { ReactComponent as TerraformLogo } from '../assets/img/terraform-logo.svg';
import WeaveGitOps from '../assets/img/weave-logo.svg';
import { Routes } from '../utils/nav';

interface SubNavItem {
  name: string;
  link: string;
  isVisible: boolean;
}
interface NavigationItem {
  icon?: any;
  name: string;
  link: string;
  subItems?: Array<SubNavItem>;
  isVisible?: boolean;
}

const NavWrapper = styled.div`
  display: flex;
  align-items: center;
  justify-content: start;
  flex-direction: column;
  margin-bottom: ${({ theme }) => theme.spacing.small};

  a.route-nav {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: start;
    padding-top: ${({ theme }) => theme.spacing.xs};
    padding-bottom: ${({ theme }) => theme.spacing.xs};
    padding-left: ${({ theme }) => theme.spacing.medium};
    padding-right: ${({ theme }) => theme.spacing.medium};
  }

  .parent-icon {
    width: 20px;
    height: 20px;
  }

  span.parent-route {
    margin-left: ${({ theme }) => theme.spacing.xs};
  }

  a:not(a.nav-link-active):hover {
    background: ${({ theme }) => theme.colors.neutral10};
  }
  .subroute-container {
    width: 100%;
  }

  .subroute-nav {
    padding: ${({ theme }) => theme.spacing.xs}
      ${({ theme }) => theme.spacing.xs} ${({ theme }) => theme.spacing.xs}
      calc(
        ${({ theme }) => theme.spacing.medium} * 2 +
          ${({ theme }) => theme.spacing.xxs}
      );
    color: ${({ theme }) => theme.colors.neutral30};
    font-weight: 600;
  }
`;

export const NavItem = styled(NavLink).attrs({
  activeClassName: 'nav-link-active',
})`
  display: flex;
  font-size: ${18}px;
  box-sizing: border-box;
  color: ${({ theme }) => theme.colors.neutral40};
  font-weight: 600;
  &.${props => props.activeClassName} {
    border-right: 3px solid ${({ theme }) => theme.colors.primary};
    background: rgba(0, 179, 236, 0.1);
    color: ${({ theme }) => theme.colors.primary};

    svg {
      fill: ${({ theme }) => theme.colors.primary};
    }
  }
`;

const useStyles = makeStyles({
  root: {
    paddingTop: theme.spacing.medium,
    alignItems: 'center',
    height: '100vh',
    borderTopRightRadius: '10px',
  },
  logo: {
    padding: `calc(${theme.spacing.medium} - ${theme.spacing.xxs})`,
  },
});

const NavItems = (navItems: Array<NavigationItem>) => {
  return navItems.map(item => {
    if (item.isVisible === false) {
      return null;
    }

    return (
      <NavWrapper key={item.name}>
        <NavItem
          exact={!!item.subItems ? true : false}
          to={item.link}
          className="route-nav"
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
                    className="subroute-nav"
                  >
                    {subItem.name}
                  </NavItem>
                )
              );
            })}
          </div>
        )}
      </NavWrapper>
    );
  });
};

export const Navigation: FC = () => {
  const { data: flagsRes } = useFeatureFlags();
  const classes = useStyles();
  const navItems: Array<NavigationItem> = [
    {
      name: 'Clusters',
      link: '/clusters',
      icon: <Clusters />,
      subItems: [
        {
          name: 'Violation Log',
          link: '/clusters/violations',
          isVisible: true,
        },
      ],
    },
    {
      name: 'Applications',
      link: V2Routes.Automations,
      icon: <Applications />,
      subItems: [
        {
          name: 'Sources',
          link: V2Routes.Sources,
          isVisible: true,
        },
        {
          name: 'Pipelines',
          link: '/applications/pipelines',
          isVisible: !!flagsRes?.flags?.WEAVE_GITOPS_FEATURE_PIPELINES,
        },
        {
          name: 'Delivery',
          link: '/applications/delivery',
          isVisible:
            process.env.REACT_APP_DISABLE_PROGRESSIVE_DELIVERY !== 'true',
        },
      ],
    },
    {
      name: 'Templates',
      link: '/templates',
      icon: <Templates />,
    },
    {
      name: 'Terraform',
      link: Routes.TerraformObjects,
      icon: <TerraformLogo />,
      isVisible: !!flagsRes?.flags?.WEAVE_GITOPS_FEATURE_TERRAFORM_UI,
    },
    {
      name: 'Flux Runtime',
      link: V2Routes.FluxRuntime,
      icon: <FluxIcon />,
    },
    {
      name: 'Policies',
      link: '/policies',
      icon: <Policies />,
    },
  ];
  return (
    <>
      <div title="Home" className={classes.logo}>
        <img src={WeaveGitOps} alt="Home" />
      </div>
      <Box className={classes.root} bgcolor={theme.colors.white}>
        {NavItems(navItems)}
      </Box>
    </>
  );
};
