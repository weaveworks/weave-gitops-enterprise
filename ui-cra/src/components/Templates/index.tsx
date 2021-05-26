import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import Template from './Card';
import { makeStyles } from '@material-ui/core/styles';
import Grid from '@material-ui/core/Grid';

const useStyles = makeStyles({
  gridContainer: {
    paddingRight: '40px',
  },
});

const TemplatesDashboard = () => {
  const classes = useStyles();
  return (
    <PageTemplate documentTitle="WeGO · Templates">
      <Grid
        container
        spacing={4}
        className={classes.gridContainer}
        justify="center"
      >
        <Grid item xs={12} sm={6} md={4}>
          <Template />
        </Grid>
        <Grid item xs={12} sm={6} md={4}>
          <Template />
        </Grid>
        <Grid item xs={12} sm={6} md={4}>
          <Template />
        </Grid>
      </Grid>
    </PageTemplate>
  );
};

export default TemplatesDashboard;
