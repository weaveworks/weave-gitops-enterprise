import { Flex, theme } from '@weaveworks/weave-gitops';
import { NavLink } from 'react-router-dom';
import styled from 'styled-components';
import { makeStyles } from '@material-ui/core/styles';

const { xs, medium } = theme.spacing;
const { neutral40, primary, neutral30 } = theme.colors;

export const useStyles = makeStyles({
    root: {
      paddingTop: medium,
      alignItems: 'center',
      height: 'calc(100vh - 84px)',
      borderTopRightRadius: '10px',
    },
    navWrapper: {
      display: 'flex',
      flexDirection: 'column',
    },
    sideNavOpened: {
      width: '210px',
      transition: 'width 0.3s ease-in-out',
    },
    sidenavClosed: {
      transition: 'width 0.3s ease-in-out',
      width: '60px',
    },
    collapseText: {
      fontSize: '12px',
      fontWeight: 700,
      color: '#737373',
    },
});
  
export const NavGroupItemWrapper = styled(Flex)`
  padding-top: ${xs};
  font-size: 12px;
  .title {
    color: ${neutral30};
    font-weight: 600;
    padding: 8px 0 8px 20px;
  }
  &:not(:first-of-type, .tilte) {
    margin-top: 8px;
  }
  .toggleOpacity {
    opacity: ${props => (!props.collapsed ? 0 : 1)};
    transition: opacity 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
  }

  .ellipsis {
    white-space: nowrap;
    overflow: hidden;
    word-wrap: normal;
  }
  a.route-nav {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: ${props => (props.collapsed ? ' start' : 'center')};
    padding: ${props => (props.collapsed ? ' 4px 0 4px 24px' : '4px 0')};
  }
  .route-item {
    width: ${props => (!props.collapsed ? 0 : 'auto')};
    margin-left: ${props => (!props.collapsed ? 0 : '12px')};
    transition: width 0.3s;
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
      stroke: ${primary};
    }
  }
`;

export const LogoWrapper = styled.div`
  display: flex;
  margin: auto;
  padding-top: ${medium};
  padding-bottom: 17px;
  .logo {
    opacity: 1;
  }
  .toggleOpacity {
    opacity: 0;
    width: 0;
    transition: opacity 0.5s;
  }
`;

export const CollapseWrapper = styled(Flex)`
  align-items: center;
  justify-content: end;
  margin: 20px 20px 0 0;
  font-size: 12px;
  color: #737373;
  font-weight: 700;
  .toggleOpacity {
    opacity: ${props => (!props.collapsed ? 0 : 1)};
    width: ${props => (!props.collapsed ? 0 : 'auto')};
    transition: opacity 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
    transition: width 0.2s;
  }
  .icon {
    font-size: 24px;
    background: #737373;
    border-radius: 50%;
    color: white;
  }
`;
