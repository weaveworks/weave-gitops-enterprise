import React, { FC } from 'react';
import styled, { css } from 'styled-components';
import { theme, V2Routes } from '@weaveworks/weave-gitops';
import { NavLink } from 'react-router-dom';
import WeaveGitOps from '../assets/img/weave-logo.svg';
import TitleLogo from '../assets/img/title.svg';
import { makeStyles } from '@material-ui/core/styles';
import Box from '@material-ui/core/Box';
import Divider from '@material-ui/core/Divider';

const itemCss = css`
  /* breaking from std. spacing as */
  display: flex;
  font-size: ${20}px;
  line-height: ${theme.spacing.large};
  height: ${theme.spacing.large};
  box-sizing: border-box;
  color: ${theme.colors.neutral40};
  font-weight: 600;
  padding: 0 ${theme.spacing.small} ${theme.spacing.small} 0;
  margin: 0 0 ${theme.spacing.small} 0;
`;

const itemActiveCss = css`
  border-right: 4px solid ${theme.colors.primary};
`;

const Title = styled.div`
  align-items: center;
  display: flex;
  color: ${theme.colors.white};
  font-size: ${20}px;
  background: ${theme.colors.primary};
  height: ${80}px;
  position: sticky;
  top: 0;
`;

const Logo = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
`;

//  How to mix NavLink and SC? Like so:
//  https://github.com/styled-components/styled-components/issues/184

export const NavItem = styled(NavLink).attrs({
  activeClassName: 'nav-link-active',
})`
  ${itemCss}

  &.${props => props.activeClassName} {
    ${itemActiveCss}
  }
`;

const useStyles = makeStyles({
  root: {
    paddingTop: theme.spacing.large,
    paddingLeft: theme.spacing.medium,
    alignItems: 'center',
    marginTop: theme.spacing.medium,
    height: '100vh',
    borderTopRightRadius: theme.spacing.xs,
  },
  subItem: {
    opacity: 0.7,
    fontWeight: 400,
  },
  section: {
    paddingBottom: theme.spacing.small,
  },
});

export const Navigation: FC = () => {
  const classes = useStyles();

  return (
    <>
      <Title title="Home">
        <Logo>
          <img
            src={WeaveGitOps}
            alt="WG-logo"
            style={{ height: 56, paddingLeft: theme.spacing.medium }}
          />
          <Divider style={{ margin: theme.spacing.xxs }} />
          <img src={TitleLogo} alt="WG-text" />
        </Logo>
      </Title>
      <Box className={classes.root} bgcolor={theme.colors.white}>
        <Box className={classes.section}>
          <NavItem to="/clusters" exact>
            Clusters
          </NavItem>
          <NavItem className={classes.subItem} to="/clusters/templates">
            Templates
          </NavItem>
          <NavItem className={classes.subItem} to="/clusters/violations">
            Violation Log
          </NavItem>
        </Box>
        <Box className={classes.section}>
          <NavItem to={V2Routes.Automations} exact>
            Applications
          </NavItem>
          <NavItem className={classes.subItem} to={V2Routes.Sources}>
            Sources
          </NavItem>
          {Boolean(process.env.REACT_APP_ENABLE_PROGRESSIVE_DELIVERY) && (
            <NavItem className={classes.subItem} to="/applications/delivery">
              Delivery
            </NavItem>
          )}
          <NavItem to={V2Routes.FluxRuntime}>Flux Runtime</NavItem>
        </Box>
        <Box className={classes.section}>
          <NavItem to="/policies">Policies</NavItem>
        </Box>
      </Box>
    </>
  );
};
