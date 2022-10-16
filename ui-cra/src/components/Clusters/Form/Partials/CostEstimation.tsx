import { createStyles, Grid, makeStyles } from '@material-ui/core';
import { Button, theme } from '@weaveworks/weave-gitops';
import React, { FC } from 'react';

const CostEstimation: FC<{
  isCostEstimationEnabled?: string;
  costEstimation?: String;
  handleCostEstimation: () => Promise<void>;
}> = ({
  isCostEstimationEnabled = 'false',
  handleCostEstimation,
  costEstimation = '00.00',
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
          <div>
            <Grid container>
              <Grid
                item
                xs={6}
                sm={6}
                md={6}
                lg={6}
                justifyContent="flex-start"
                alignItems="center"
                container
              >
                <div>Monthly Cost:</div>
              </Grid>
              <Grid
                item
                xs={6}
                sm={6}
                md={6}
                lg={6}
                justifyContent="flex-end"
                alignItems="center"
                container
              >
                <div className={classes.costWrapper}>{costEstimation}</div>
              </Grid>
            </Grid>
          </div>
        </Grid>
        <Grid
          item
          xs={6}
          sm={6}
          md={6}
          lg={6}
          justifyContent="flex-end"
          alignItems="center"
          container
        >
          <div>
            <Button
              id="get-estimation"
              className={classes.getEstimationButton}
              onClick={() => handleCostEstimation()}
            >
              GET ESTIMATION
            </Button>
          </div>
        </Grid>
      </Grid>
    </div>
  );
};

export default CostEstimation;
