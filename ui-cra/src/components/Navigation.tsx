import React, { FC } from 'react';
import styled, { css } from 'styled-components';
import theme from 'weaveworks-ui-components/lib/theme';
import { NavLink } from 'react-router-dom';
import WeaveLogo from '../assets/img/wego.svg';
import { makeStyles } from '@material-ui/core/styles';

const Bar = styled.div`
  background-color: ${theme.colors.white};
  box-shadow: ${theme.boxShadow.light};
  padding: 20px;
`;

const Content = styled.div`
  align-items: center;
`;

const navItemPadding = css`
  padding: 0 ${theme.spacing.small};
`;

const itemCss = css`
  /* breaking from std. spacing as */
  display: flex;
  font-size: ${20}px;
  line-height: ${theme.spacing.xxl};
  height: ${theme.spacing.xxl};
  box-sizing: border-box;
  color: ${theme.colors.black};
  overflow: auto;
  ${navItemPadding}
`;

const itemActiveCss = css`
  border-left: 4px solid ${theme.colors.blue400};
`;

const Title = styled.a`
  display: flex;
  justify-content: center;
`;

const Logo = styled.div`
  width: 80px;
  height: 80px;
  background: url(${WeaveLogo});
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

export const FlexSpacer = styled.div`
  flex: 1;
`;

const useStyles = makeStyles({
  root: {
    height: '100vh',
  },
});

export const Navigation: FC = () => {
  const classes = useStyles();

  return (
    <Bar className={classes.root}>
      <Content>
        <Title title="Home">
          <Logo />
        </Title>
        <NavItem to="/clusters">Clusters</NavItem>
        <NavItem to="/templates">Templates</NavItem>
        <NavItem to="/alerts">Alerts</NavItem>
      </Content>
    </Bar>
  );
};
