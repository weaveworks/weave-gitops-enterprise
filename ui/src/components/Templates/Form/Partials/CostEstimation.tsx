import { createStyles, Grid, makeStyles } from '@material-ui/core';
import { Alert } from '@material-ui/lab';
import { Button } from '@weaveworks/weave-gitops';
import React, { Dispatch, FC } from 'react';

const CostEstimation: FC<{
  costEstimate: string;
  isCostEstimationLoading: boolean;
  costEstimateMessage: string;
  handleCostEstimation: () => Promise<void>;
  setFormError: Dispatch<React.SetStateAction<string>>;
  setSubmitType: Dispatch<React.SetStateAction<string>>;
}> = ({
  costEstimate,
  isCostEstimationLoading,
  costEstimateMessage,
  setSubmitType,
}) => {
  const useStyles = makeStyles(() =>
    createStyles({
      costWrapper: {
        marginRight: '12px',
        fontWeight: 'bold',
      },
      errorMessage: {
        marginTop: '8px',
        marginRight: '12px',
        borderRadius: '4px',
      },
    }),
  );
  const classes = useStyles();
  return (
    <>
      <h2>Cost Estimation</h2>
      <Grid alignItems="center" container style={{ paddingRight: '24px' }}>
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
            loading={isCostEstimationLoading}
            disabled={isCostEstimationLoading}
            id="get-estimation"
            type="submit"
            onClick={() => setSubmitType('Get cost estimation')}
          >
            GET ESTIMATION
          </Button>
        </Grid>
        {costEstimateMessage && (
          <Alert className={classes.errorMessage} severity="warning">
            {costEstimateMessage}
          </Alert>
        )}
      </Grid>
    </>
  );
};

export default CostEstimation;
