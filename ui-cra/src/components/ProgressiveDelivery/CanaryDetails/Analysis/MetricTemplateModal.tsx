import { usePolicyStyle } from '../../../Policies/PolicyStyles';

import Dialog from '@material-ui/core/Dialog';
import { CanaryMetricTemplate } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import React, { Dispatch, FC } from 'react';

import { DialogContent, DialogTitle } from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { darcula } from 'react-syntax-highlighter/dist/cjs/styles/prism';
import { CloseIconButton } from '../../../../assets/img/close-icon-button';

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
          <SyntaxHighlighter
            language="yaml"
            style={darcula}
            wrapLongLines="pre-wrap"
            showLineNumbers={true}
            codeTagProps={{
              className: classes.code,
            }}
            customStyle={{
              height: '450px',
            }}
          >
            {metricTemplate.yaml}
          </SyntaxHighlighter>
        </DialogContent>
      </Dialog>
    </div>
  );
};
