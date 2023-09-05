import Dialog from '@material-ui/core/Dialog';
import { CanaryMetricTemplate } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import React, { Dispatch, FC } from 'react';

import { DialogContent } from '@material-ui/core';
import { Flex, YamlView } from '@weaveworks/weave-gitops';
import { MuiDialogTitle } from '../../../Shared';

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
  return (
    <Flex wide>
      <Dialog
        id="metric-template-dialog"
        open={open}
        maxWidth="md"
        fullWidth
        scroll="paper"
      >
        <MuiDialogTitle
          title={`Metric Template: ${metricTemplate.name}`}
          onFinish={() => setOpenMetricTemplate(false)}
        />
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
    </Flex>
  );
};
