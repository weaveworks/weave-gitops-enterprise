import { CanaryMetric } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { AppContext, DataTable, Text } from '@weaveworks/weave-gitops';
import React, { useContext } from 'react';
import { TableWrapper } from '../../../Shared';

export const CanaryMetricsTable = ({
  metrics,
}: {
  metrics: CanaryMetric[];
}) => {
  const { setDetailModal } = useContext(AppContext);
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
                <Text
                  onClick={() => {
                    const metricObj: any = {
                      ...c.metricTemplate,
                      type: 'MetricTemplate',
                    };
                    setDetailModal({
                      object: metricObj,
                    });
                  }}
                  color="primary10"
                  pointer
                >
                  {c.metricTemplate?.name}
                </Text>
              ) : (
                ''
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
