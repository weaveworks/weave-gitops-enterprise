import React, { FC } from 'react';
import styled, { css } from 'styled-components';
import theme from 'weaveworks-ui-components/lib/theme';
import { NavLink } from 'react-router-dom';
import WeaveGitOps from '../assets/img/weave-logo.svg';
import { makeStyles } from '@material-ui/core/styles';
import Box from '@material-ui/core/Box';

const navItemPadding = css`
  padding: 0 ${theme.spacing.small} ${theme.spacing.small}
    ${theme.spacing.small};
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
  border-left: 4px solid ${theme.colors.blue400};
`;

const Title = styled.div`
  display: flex;
  justify-content: center;
  padding-top: ${theme.spacing.large};
  background: #00b3ec;
  max-height: 112px;
`;

const Logo = styled.div`
  width: 130px;
  height: 112px;
  background: url(${WeaveGitOps});
  background-repeat: no-repeat;
`;

//  How to mix NavLink and SC? Like so:
//  https://github.com/styled-components/styled-components/issues/184

const NavItem = styled(NavLink).attrs({
  activeClassName: 'nav-link-active',
})`
  ${itemCss}

  &.${props => props.activeClassName} {
    ${itemActiveCss}
  }
`;

const useStyles = makeStyles({
  root: {
    padding: theme.spacing.medium,
    alignItems: 'center',
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
        <Logo />
      </Title>
      <Box className={classes.root} bgcolor={theme.colors.white}>
        <Box className={classes.section}>
          <NavItem className={classes.bold} to="/clusters">
            Clusters
          </NavItem>
          <NavItem to="/templates">Templates</NavItem>
          <NavItem to="/alerts">Alerts</NavItem>
        </Box>
        <Box className={classes.section}>
          <NavItem className={classes.bold} to="/applications">
            Applications
          </NavItem>
        </Box>
      </Box>
    </>
  );
};
