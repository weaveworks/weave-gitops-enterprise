import { theme, FilterableTable, SortType } from '@weaveworks/weave-gitops';
import { ThemeProvider } from 'styled-components';
import { usePolicyStyle } from '../../../Policies/PolicyStyles';
import { TableWrapper } from '../../CanaryStyles';
import {
  CanaryMetric,
  CanaryMetricTemplate,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import React, { Dispatch, FC, SetStateAction, useState } from 'react';
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
  metrics: CanaryMetric[];
}) => {
  const classes = usePolicyStyle();

  const [openMetricTemplate, setOpenMetricTemplate] = useState(false);

  console.log(openMetricTemplate);
  return (
    <div className={classes.root}>
      <ThemeProvider theme={theme}>
        {metrics.length > 0 ? (
          <TableWrapper id="canary-analysis-metrics">
            <FilterableTable
              filters={{}}
              rows={metrics}
              fields={[
                {
                  label: 'Name',
                  value: 'name',
                  textSearchable: true,
                  sortType: SortType.string,
                  sortValue: ({ name }) => name,
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
                  textSearchable: true,
                  sortType: SortType.string,
                  sortValue: ({ namespace }) => namespace,
                },
                {
                  label: 'Threshold Min',
                  value: (c: CanaryMetric) =>
                    c.thresholdRange?.min ? '' + c.thresholdRange?.min : '-',
                  sortValue: ({ min }) => min || 0,
                },
                {
                  label: 'Threshold Max',
                  value: (c: CanaryMetric) =>
                    c.thresholdRange?.max ? '' + c.thresholdRange?.max : '-',
                  sortValue: ({ max }) => max || 0,
                },
                {
                  label: 'Interval',
                  value: 'interval',
                },
              ]}
            />
          </TableWrapper>
        ) : (
          <p>No metrics to display</p>
        )}
      </ThemeProvider>
    </div>
  );
};
