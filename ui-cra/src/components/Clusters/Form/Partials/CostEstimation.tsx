import { createStyles, Grid, makeStyles } from '@material-ui/core';
import { Button, theme } from '@weaveworks/weave-gitops';
import React, { FC, useEffect } from 'react';

const costFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
});

const CostEstimation: FC<{
  isCostEstimationEnabled?: string;
  costEstimation?: any;
  handleCostEstimation: () => Promise<void>;
}> = ({
  isCostEstimationEnabled = 'false',
  handleCostEstimation,
  costEstimation,
}) => {
  const [estimate, setEstimate] = React.useState('$0.00 USD');
  useEffect(() => {
    if (costEstimation) {
      const estimate =
        costEstimation?.amount !== undefined
          ? `${costFormatter.format(costEstimation.amount)} ${
              costEstimation.currency
            }`
          : `${costFormatter.format(
              costEstimation.range.low,
            )} - ${costFormatter.format(costEstimation.range.high)} ${
              costEstimation.currency
            }`;
      setEstimate(estimate);
    }
  }, [costEstimation]);

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
                <div className={classes.costWrapper}>{estimate}</div>
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
