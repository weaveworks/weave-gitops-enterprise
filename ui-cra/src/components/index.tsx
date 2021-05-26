import React, { FC } from 'react';
import Grid from '@material-ui/core/Grid';
import { Switch, Route } from 'react-router-dom';
import ClustersProvider from '../contexts/Clusters/Provider';
import AlertsProvider from '../contexts/Alerts/Provider';
import MCCP from './Clusters';
import TemplatesDashboard from './Templates';
import { Navigation } from './Navigation';
import { AlertsDashboard } from './Alerts';

export const PageSwitch: FC = () => {
  return (
    <ClustersProvider>
      <AlertsProvider>
        <Grid container>
          <Grid item xs={2}>
            <Navigation />
          </Grid>
          <Grid item xs={10}>
            <Switch>
              <Route component={MCCP} exact={true} path="/clusters" />
              <Route
                component={TemplatesDashboard}
                exact={true}
                path="/templates"
              />
              <Route component={AlertsDashboard} exact={true} path="/alerts" />
            </Switch>
          </Grid>
        </Grid>
      </AlertsProvider>
    </ClustersProvider>
  );
};
