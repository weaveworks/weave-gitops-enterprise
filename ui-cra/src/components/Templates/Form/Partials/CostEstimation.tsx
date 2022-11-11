import { createStyles, Grid, makeStyles } from '@material-ui/core';
import { Alert } from '@material-ui/lab';
import { Button, LoadingPage, theme } from '@weaveworks/weave-gitops';
import React, { Dispatch, FC } from 'react';
import { validateFormData } from '../../../../utils/form';

const CostEstimation: FC<{
  costEstimate: string;
  isCostEstimationLoading: boolean;
  costEstimateMessage: string;
  handleCostEstimation: () => Promise<void>;
  setFormError: Dispatch<React.SetStateAction<any>>;
}> = ({
  handleCostEstimation,
  costEstimate,
  isCostEstimationLoading,
  costEstimateMessage,
  setFormError,
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
      previewLoading: {
        padding: theme.spacing.base,
      },
      errorMessage: {
        marginTop: theme.spacing.xs,
        marginRight: theme.spacing.medium,
        borderRadius: theme.spacing.xxs,
      },
    }),
  );
  const classes = useStyles();
  return (
    <>
      <h2>Cost Estimation</h2>
      {isCostEstimationLoading ? (
        <LoadingPage className={classes.previewLoading} />
      ) : (
        <Grid alignItems="center" container>
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
              onClick={event =>
                validateFormData(event, handleCostEstimation, setFormError)
              }
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
      )}
    </>
  );
};

export default CostEstimation;
