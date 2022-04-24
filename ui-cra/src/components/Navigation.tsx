import React, { FC } from 'react';
import styled, { css } from 'styled-components';
import { theme, V2Routes } from '@weaveworks/weave-gitops';
import { NavLink } from 'react-router-dom';
import WeaveGitOps from '../assets/img/weave-logo.svg';
import TitleLogo from '../assets/img/title.svg';
import { makeStyles } from '@material-ui/core/styles';
import Box from '@material-ui/core/Box';
import Divider from '@material-ui/core/Divider';

const navItemPadding = css`
  padding: 0 ${theme.spacing.small} ${theme.spacing.small} 0;
`;

const navItemMargin = css`
  margin: 0 0 ${theme.spacing.small} 0;
`;

const itemCss = css`
  /* breaking from std. spacing as */
  display: flex;
  font-size: ${20}px;
  line-height: ${theme.spacing.large};
  height: ${theme.spacing.large};
  box-sizing: border-box;
  color: ${theme.colors.black};
  ${navItemPadding}
  ${navItemMargin}
`;

const itemActiveCss = css`
  border-right: 4px solid ${theme.colors.primary};
`;

const Title = styled.div`
  align-items: center;
  display: flex;
  color: ${theme.colors.white};
  font-size: ${20}px;
  padding-left: ${theme.spacing.medium};
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
  bold: {
    fontWeight: 600,
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
          <img src={WeaveGitOps} alt="logo-image" style={{ height: 56 }} />
          <Divider style={{ margin: theme.spacing.xxs }} />
          <img src={TitleLogo} alt="logo-text" />
        </Logo>
      </Title>
      <Box className={classes.root} bgcolor={theme.colors.white}>
        <Box className={classes.section}>
          <NavItem className={classes.bold} to="/clusters" exact>
            Clusters
          </NavItem>
          <NavItem to="/clusters/templates">Templates</NavItem>
          <NavItem to="/clusters/alerts" exact>
            Alerts
          </NavItem>
        </Box>
        <Box className={classes.section}>
          <NavItem className={classes.bold} to={V2Routes.Automations}>
            Applications
          </NavItem>
          <NavItem to={V2Routes.Sources}>Sources</NavItem>
          <NavItem to={V2Routes.FluxRuntime}>Flux Runtime</NavItem>
        </Box>
        <Box className={classes.section}>
          <NavItem className={classes.bold} to="/policies">
            Policies
          </NavItem>
        </Box>
      </Box>
    </>
  );
};
