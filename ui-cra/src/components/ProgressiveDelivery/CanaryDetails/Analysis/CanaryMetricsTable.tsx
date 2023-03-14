import {
  CanaryMetric,
  CanaryMetricTemplate,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { DataTable } from '@weaveworks/weave-gitops';
import { Dispatch, FC, SetStateAction, useState } from 'react';
import { TableWrapper } from '../../../Shared';
import { MetricTemplateModal } from './MetricTemplateModal';

type Props = {
  metricTemplate: CanaryMetricTemplate;
  openMetricTemplate: boolean;
  setOpenMetricTemplate: Dispatch<SetStateAction<any>>;
};

const MetricTemplateModalWrapper: FC<Props> = ({
  metricTemplate,
  openMetricTemplate,
  setOpenMetricTemplate,
}) => {
  return (
    <div>
      <a
        target="_self"
        href="#metric-template-dialog"
        onClick={() => setOpenMetricTemplate(true)}
      >
        {metricTemplate.name}
      </a>
      <MetricTemplateModal
        open={openMetricTemplate}
        metricTemplate={metricTemplate}
        setOpenMetricTemplate={setOpenMetricTemplate}
      />
    </div>
  );
};

export const CanaryMetricsTable = ({
  metrics,
}: {
  metrics?: CanaryMetric[];
}) => {
  const [openMetricTemplate, setOpenMetricTemplate] = useState(false);

  return (
    <TableWrapper id="canary-analysis-metrics">
      <DataTable
        rows={metrics}
        fields={[
          {
            label: 'Name',
            value: 'name',
          },
          {
            label: 'Metric Template',
            value: (c: CanaryMetric) =>
              c.metricTemplate ? (
                <MetricTemplateModalWrapper
                  metricTemplate={c.metricTemplate}
                  openMetricTemplate={openMetricTemplate}
                  setOpenMetricTemplate={setOpenMetricTemplate}
                />
              ) : (
                '-'
              ),
          },
          {
            label: 'Threshold Min',
            value: (c: CanaryMetric) =>
              c.thresholdRange?.min ? '' + c.thresholdRange?.min : '-',
          },
          {
            label: 'Threshold Max',
            value: (c: CanaryMetric) =>
              c.thresholdRange?.max ? '' + c.thresholdRange?.max : '-',
          },
          {
            label: 'Interval',
            value: 'interval',
          },
        ]}
      />
    </TableWrapper>
  );
};
