import React from 'react';
import { Switch, Route } from 'react-router-dom';
import ClustersProvider from '../contexts/Clusters/Provider';
import AlertsProvider from '../contexts/Alerts/Provider';
import MCCP from './Clusters';
import TemplatesDashboard from './Templates';
import { Navigation } from './Navigation';
import { AlertsDashboard } from './Alerts';
import CssBaseline from '@material-ui/core/CssBaseline';
import Drawer from '@material-ui/core/Drawer';
import Hidden from '@material-ui/core/Hidden';
import IconButton from '@material-ui/core/IconButton';
import { ReactComponent as MenuIcon } from '../assets/img/menu-burger.svg';
import {
  makeStyles,
  useTheme,
  Theme,
  createStyles,
} from '@material-ui/core/styles';
import AddCluster from './Clusters/Create';
import TemplatesProvider from '../contexts/Templates/Provider';
import Box from '@material-ui/core/Box';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import Compose from './ProvidersCompose';

const drawerWidth = 240;

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      display: 'flex',
    },
    drawer: {
      [theme.breakpoints.up('sm')]: {
        width: drawerWidth,
        flexShrink: 0,
      },
    },
    menuButton: {
      [theme.breakpoints.up('sm')]: {
        display: 'none',
      },
      marginLeft: weaveTheme.spacing.small,
    },
    menuButtonBox: {
      padding: weaveTheme.spacing.medium,
      alignItems: 'center',
    },
    drawerPaper: {
      width: drawerWidth,
    },
    content: {
      flexGrow: 1,
      padding: weaveTheme.spacing.small,
    },
  }),
);

const ResponsiveDrawer = () => {
  const classes = useStyles();
  const theme = useTheme();
  const [mobileOpen, setMobileOpen] = React.useState(false);

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  return (
    <Compose components={[ClustersProvider, TemplatesProvider, AlertsProvider]}>
      <div className={classes.root}>
        <CssBaseline />
        <nav className={classes.drawer} aria-label="mailbox folders">
          <Hidden smUp implementation="css">
            <Box className={classes.menuButtonBox}>
              <IconButton
                color="inherit"
                aria-label="open drawer"
                edge="start"
                onClick={handleDrawerToggle}
                className={classes.menuButton}
              >
                <MenuIcon />
              </IconButton>
            </Box>
            <Drawer
              variant="temporary"
              anchor={theme.direction === 'rtl' ? 'right' : 'left'}
              open={mobileOpen}
              onClose={handleDrawerToggle}
              classes={{
                paper: classes.drawerPaper,
              }}
              ModalProps={{
                keepMounted: true, // Better open performance on mobile.
              }}
            >
              <Navigation />;
            </Drawer>
          </Hidden>
          <Hidden xsDown implementation="css">
            <Drawer
              classes={{
                paper: classes.drawerPaper,
              }}
              variant="permanent"
              open
            >
              <Navigation />;
            </Drawer>
          </Hidden>
        </nav>
        <main className={classes.content}>
          <Switch>
            <Route component={MCCP} exact path={['/', '/clusters']} />
            <Route
              component={AddCluster}
              exact
              path="/templates/:templateName/create"
            />
            <Route component={TemplatesDashboard} exact path="/templates" />
            <Route component={AlertsDashboard} exact path="/alerts" />
          </Switch>
        </main>
      </div>
    </Compose>
  );
};

export default ResponsiveDrawer;
