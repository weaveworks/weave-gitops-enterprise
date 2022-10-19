import { createStyles, Grid, makeStyles } from '@material-ui/core';
import { Button, theme } from '@weaveworks/weave-gitops';
import React, { FC, useEffect } from 'react';
import { validateFormData } from '../../../../utils/form';

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
      const {
        currency,
        amount,
        range: { low, high },
      } = costEstimation;
      const estimate =
        amount !== undefined
          ? `${costFormatter.format(amount)} ${currency}`
          : `${costFormatter.format(low)} - ${costFormatter.format(
              high,
            )} ${currency}`;
      setEstimate(estimate);
    } else {
      setEstimate('$0.00 USD');
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
              <div className={classes.costWrapper}>{estimate}</div>
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
