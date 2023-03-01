import Box from '@material-ui/core/Box';
import { Link, theme } from '@weaveworks/weave-gitops';
import React, { FC, useState } from 'react';

import { IconButton } from '@material-ui/core';
import { ArrowLeft, ArrowRight } from '@material-ui/icons';
import LogoIcon from '../../assets/img/logo.svg';
import WeaveGitOps from '../../assets/img/weave-logo.svg';
import { Routes } from '../../utils/nav';
import NavItems from './NavItems';
import { CollapseWrapper, LogoWrapper, useNavStyles } from './styles';

const MemoizedNavItems = React.memo(NavItems);

export const Navigation: FC = () => {
  const classes = useNavStyles();
  const [collapsed, setCollapsed] = useState(true);

  return (
    <div
      className={`${classes.navWrapper} ${
        collapsed ? classes.sideNavOpened : classes.sidenavClosed
      }`}
    >
      <LogoWrapper title="Home">
        <Link to={Routes.Clusters}>
          <img
            src={WeaveGitOps}
            alt="Home"
            className={collapsed ? 'logo' : 'toggleOpacity'}
          />
          <img
            src={LogoIcon}
            alt="Home"
            className={!collapsed ? 'logo' : 'toggleOpacity'}
          />
        </Link>
      </LogoWrapper>
      <Box
        className={`nav-items ${classes.root} `}
        bgcolor={theme.colors.white}
      >
        <div>
          <MemoizedNavItems collapsed={collapsed} />
        </div>
        <CollapseWrapper collapsed={collapsed}>
          <div className="toggleOpacity">Collapse</div>
          <IconButton size="small" onClick={() => setCollapsed(!collapsed)}>
            {collapsed ? (
              <ArrowLeft className="icon" />
            ) : (
              <ArrowRight className="icon" />
            )}
          </IconButton>
        </CollapseWrapper>
      </Box>
    </div>
  );
};
