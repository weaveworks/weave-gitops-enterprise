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
import {
  AppContextProvider,
  applicationsClient,
} from '@weaveworks/weave-gitops';
import TemplatesProvider from '../contexts/Templates/Provider';
import NotificationsProvider from '../contexts/Notifications/Provider';
import Compose from './ProvidersCompose';
import Box from '@material-ui/core/Box';
import { PageTemplate } from './Layout/PageTemplate';
import { SectionHeader } from './Layout/SectionHeader';
import { ContentWrapper } from './Layout/ContentWrapper';
import Lottie from 'react-lottie-player';
import error404 from '../assets/img/error404.json';
import AddClusterWithCredentials from './Clusters/Create';
import WGApplicationsDashboard from './Applications';
import WGApplicationDetail from './Applications/Detail';

const APPS_ROUTE = '/applications';
const APP_DETAIL_ROTUE = '/application_detail';

const drawerWidth = 220;

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
    appBar: {
      [theme.breakpoints.up('sm')]: {
        width: `calc(100% - ${drawerWidth}px)`,
        marginLeft: drawerWidth,
      },
      backgroundColor: '#00b3ec',
      boxShadow: 'none',
    },
    menuButton: {
      [theme.breakpoints.up('sm')]: {
        display: 'none',
      },
      marginLeft: 0,
    },
    menuButtonBox: {
      height: '80px',
      background: '#00b3ec',
      display: 'flex',
      justifyContent: 'center',
      position: 'sticky',
      top: '0',
      zIndex: 2,
    },
    toolbar: theme.mixins.toolbar,
    drawerPaper: {
      overflowY: 'hidden',
      width: drawerWidth,
      border: 'none',
      background: '#F5F5F5',
    },
    content: {
      flexGrow: 1,
    },
  }),
);

export const WGAppProvider: React.FC = props => (
  <AppContextProvider applicationsClient={applicationsClient} {...props} />
);

const ResponsiveDrawer = () => {
  const classes = useStyles();
  const theme = useTheme();
  const [mobileOpen, setMobileOpen] = React.useState(false);

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const handle404 = () => (
    <PageTemplate documentTitle="WeGO · NotFound">
      <SectionHeader />
      <ContentWrapper>
        <Lottie
          loop
          animationData={error404}
          play
          style={{ width: '100%', height: 650 }}
        />
      </ContentWrapper>
    </PageTemplate>
  );

  return (
    <Compose
      components={[
        NotificationsProvider,
        TemplatesProvider,
        ClustersProvider,
        AlertsProvider,
        WGAppProvider,
      ]}
    >
      <div className={classes.root}>
        <CssBaseline />
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
        <nav className={classes.drawer} aria-label="mailbox folders">
          <Hidden smUp implementation="css">
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
              <div className={classes.toolbar}>
                <Navigation />
              </div>
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
              <div className={classes.toolbar}>
                <Navigation />
              </div>
            </Drawer>
          </Hidden>
        </nav>
        <main className={classes.content}>
          <Switch>
            <Route component={MCCP} exact path={['/', '/clusters']} />
            <Route
              component={AddClusterWithCredentials}
              exact
              path="/clusters/templates/:templateName/create"
            />
            <Route
              component={TemplatesDashboard}
              exact
              path="/clusters/templates"
            />
            <Route component={AlertsDashboard} exact path="/clusters/alerts" />
            <Route
              component={WGApplicationsDashboard}
              exact
              path={APPS_ROUTE}
            />
            <Route
              exact
              path={APP_DETAIL_ROTUE}
              component={WGApplicationDetail}
            />
            <Route render={handle404} />
          </Switch>
        </main>
      </div>
    </Compose>
  );
};

export default ResponsiveDrawer;
