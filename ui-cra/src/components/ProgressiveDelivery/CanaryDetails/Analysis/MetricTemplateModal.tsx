import { usePolicyStyle } from '../../../Policies/PolicyStyles';

import Dialog from '@material-ui/core/Dialog';
import { CanaryMetricTemplate } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import React, { Dispatch, FC } from 'react';

import { DialogContent, DialogTitle } from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import { CloseIconButton } from '../../../../assets/img/close-icon-button';
import { YamlView } from '@weaveworks/weave-gitops';

type Props = {
  open: boolean;
  metricTemplate: CanaryMetricTemplate;
  setOpenMetricTemplate: Dispatch<React.SetStateAction<boolean>>;
};

export const MetricTemplateModal: FC<Props> = ({
  open,
  metricTemplate,
  setOpenMetricTemplate,
}) => {
  const classes = usePolicyStyle();
  return (
    <div className={classes.root}>
      <Dialog
        id="metric-template-dialog"
        open={open}
        maxWidth="md"
        fullWidth
        scroll="paper"
      >
        <DialogTitle disableTypography>
          <Typography variant="h5">
            Metric Template: {metricTemplate.name}
          </Typography>
          <CloseIconButton onClick={() => setOpenMetricTemplate(false)} />
        </DialogTitle>

        <DialogContent>
          <YamlView
            yaml={metricTemplate.yaml || ''}
            object={{
              kind: 'MetricTemplate',
              name: metricTemplate?.name,
              namespace: metricTemplate?.namespace,
            }}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
};
