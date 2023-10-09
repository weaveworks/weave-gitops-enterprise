import { createStyles, Grid, makeStyles } from '@material-ui/core';
import { Alert } from '@material-ui/lab';
import { Button } from '@weaveworks/weave-gitops';
import React, { Dispatch, FC, useCallback, useEffect, useState } from 'react';
import { TemplateEnriched } from '../../../../types/custom';
import {
  Kustomization,
  ProfileValues,
  Credential,
} from '../../../../cluster-services/cluster_services.pb';
import useNotifications from '../../../../contexts/Notifications';
import useTemplates from '../../../../hooks/templates';
import { validateFormData } from '../../../../utils/form';
import { getFormattedCostEstimate } from '../../../../utils/formatters';

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

const CostEstimation: FC<{
  template: TemplateEnriched;
  formData: any;
  profiles: ProfileValues[];
  credentials: Credential | undefined;
  kustomizations: Kustomization[];
  setFormError: Dispatch<React.SetStateAction<string>>;
}> = ({
  template,
  formData,
  profiles,
  credentials,
  kustomizations,
  setFormError,
}) => {
  const classes = useStyles();
  const { setNotifications } = useNotifications();
  const { renderTemplate } = useTemplates();
  const [costEstimate, setCostEstimate] = useState<string>('00.00 USD');
  const [costEstimateMessage, setCostEstimateMessage] = useState<string>('');
  const [costEstimationLoading, setCostEstimationLoading] =
    useState<boolean>(false);

  const handleCostEstimation = useCallback(() => {
    const { parameterValues } = formData;
    setCostEstimationLoading(true);
    return renderTemplate({
      templateName: template.name,
      templateNamespace: template.namespace,
      values: parameterValues,
      profiles,
      credentials,
      kustomizations,
      templateKind: template.templateKind,
    })
      .then(data => {
        const { costEstimate } = data;
        setCostEstimate(getFormattedCostEstimate(costEstimate));
        setCostEstimateMessage(costEstimate?.message || '');
      })
      .catch(err =>
        setNotifications([
          {
            message: { text: err.message },
            severity: 'error',
            display: 'bottom',
          },
        ]),
      )
      .finally(() => setCostEstimationLoading(false));
  }, [
    formData,
    renderTemplate,
    template.name,
    template.templateKind,
    template.namespace,
    setNotifications,
    profiles,
    credentials,
    kustomizations,
  ]);

  useEffect(() => {
    setCostEstimate('00.00 USD');
  }, [formData.parameterValues]);

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
            id="get-estimation"
            loading={costEstimationLoading}
            disabled={costEstimationLoading}
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
    </>
  );
};

export default CostEstimation;
