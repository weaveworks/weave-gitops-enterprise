import { createStyles, Grid, makeStyles } from '@material-ui/core';
import { Button, theme } from '@weaveworks/weave-gitops';
import React, { FC } from 'react';
import { validateFormData } from '../../../../utils/form';

const CostEstimation: FC<{
  isCostEstimationEnabled?: string;
  costEstimate: string;
  handleCostEstimation: () => Promise<void>;
}> = ({
  isCostEstimationEnabled = 'false',
  handleCostEstimation,
  costEstimate,
}) => {
  const useStyles = makeStyles(() =>
    createStyles({
      getEstimationButton: {
        marginRight: theme.spacing.medium,
      },
      costWrapper: {
        marginRight: theme.spacing.medium,
        fontWeight: 'bold',
      },
    }),
  );
  const classes = useStyles();
  return isCostEstimationEnabled === 'false' ? null : (
    <div>
      <h2>Cost Estimation</h2>
      <Grid container>
        <Grid item xs={6} sm={6} md={6} lg={6}>
          <Grid container>
            <Grid
              item
              xs={6}
              justifyContent="flex-start"
              alignItems="center"
              container
            >
              <div>Monthly Cost:</div>
            </Grid>
            <Grid
              item
              xs={6}
              justifyContent="flex-end"
              alignItems="center"
              container
            >
              <div className={classes.costWrapper}>{costEstimate}</div>
            </Grid>
          </Grid>
        </Grid>
        <Grid
          item
          xs={6}
          justifyContent="flex-end"
          alignItems="center"
          container
        >
          <Button
            id="get-estimation"
            className={classes.getEstimationButton}
            onClick={event => validateFormData(event, handleCostEstimation)}
          >
            GET ESTIMATION
          </Button>
        </Grid>
      </Grid>
    </div>
  );
};

export default CostEstimation;
