import { act, render, screen } from '@testing-library/react';
import { GetCanaryResponse } from '@weaveworks/progressive-delivery';
import { CanaryMetric } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import _ from 'lodash';
import { ProgressiveDeliveryProvider } from '../../../../contexts/ProgressiveDelivery';
import {
  defaultContexts,
  findTextByHeading,
  ProgressiveDeliveryMock,
  withContext,
} from '../../../../utils/test-utils';
import { CanaryMetricsTable } from '../Analysis/CanaryMetricsTable';

describe('CanaryMetricsTable', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: ProgressiveDeliveryMock;

  beforeEach(() => {
    api = new ProgressiveDeliveryMock();
    wrap = withContext([
      ...defaultContexts(),
      [ProgressiveDeliveryProvider, { api }],
    ]);
    api.IsFlaggerAvailableReturns = { clusters: { 'my-cluster': true } };
  });
  it('renders metrics table for a canary with metrics', async () => {
    let canaryAsJson = `
    {
      "namespace": "canary",
      "name": "canary01",
      "clusterName": "Default",
      "analysis": {
        "interval": "1m",
        "maxWeight": 50,
        "stepWeight": 10,
        "threshold": 5,
        "metrics": [
          {
            "name": "request-success-rate",
            "thresholdRange": {
              "min": 99,
              "max": 500        
            },
            "interval": "1m"
          }
        ]
      }
    }`;
    let canary = JSON.parse(canaryAsJson);
    api.GetCanaryReturns = {
      canary: canary,
    };

    await act(async () => {
      const c = wrap(
        <CanaryMetricsTable
          metrics={api.GetCanaryReturns.canary?.analysis?.metrics}
        />,
      );
      render(c);
    });

    expect(await screen.findByText('request-success-rate')).toBeTruthy();
    const tbl = document.querySelector('#canary-analysis-metrics table');
    const rows = tbl?.querySelectorAll('tbody tr');
    expect(rows).toHaveLength(1);
    const metric: CanaryMetric | null = _.get(api.GetCanaryReturns, [
      'canary',
      'analysis',
      'metrics',
      0,
    ]);

    if (!tbl) {
      throw new Error('Table not found');
    }

    if (!rows) {
      throw new Error('Rows not found');
    }

    if (!metric) {
      throw new Error('Metric not found');
    }

    assertCanaryMetric(tbl, rows?.item(0), metric);
  });
  it('renders metrics table for a canary with metrics with metric templates', async () => {
    const canaryAsJson = `
    {
      "namespace": "canary",
      "name": "canary01",
      "clusterName": "Default",
      "analysis": {
        "interval": "1m",
        "maxWeight": 50,
        "stepWeight": 10,
        "threshold": 5,
        "metrics": [
          {
            "name": "404s percentage",
            "thresholdRange": {
              "max": 5
            },
            "interval": "1m",
            "metricTemplate": {
              "clusterName": "Default",
              "name": "not-found-percentage",
              "namespace": "canary",
              "provider": {
                "type": "prometheus",
                "address": "http://prometheus.istio-system:9090"
              },
              "query": "100 - sum(\\n    rate(\\n        istio_requests_total{\\n          reporter=\\"destination\\",\\n          destination_workload_namespace=\\"{{ namespace }}\\",\\n          destination_workload=\\"{{ target }}\\",\\n          response_code!=\\"404\\"\\n        }[{{ interval }}]\\n    )\\n)\\n/\\nsum(\\n    rate(\\n        istio_requests_total{\\n          reporter=\\"destination\\",\\n          destination_workload_namespace=\\"{{ namespace }}\\",\\n          destination_workload=\\"{{ target }}\\"\\n        }[{{ interval }}]\\n    )\\n) * 100\\n"
            }
          }       
        ]
      }
    }`;
    const canary = JSON.parse(canaryAsJson);
    api.GetCanaryReturns = {
      canary: canary,
    };

    await act(async () => {
      const c = wrap(
        <CanaryMetricsTable
          metrics={api.GetCanaryReturns.canary?.analysis?.metrics}
        />,
      );
      render(c);
    });

    expect(await screen.findByText('404s percentage')).toBeTruthy();
    const [tbl, rows] = getTableAndRows('#canary-analysis-metrics table');

    expect(rows).toHaveLength(1);
    const metric = getCanaryMetric(api.GetCanaryReturns);

    assertCanaryMetric(tbl, rows.item(0), metric);
  });
});

function assertCanaryMetric(
  table: Element,
  metricAsElement: Element,
  metric: CanaryMetric,
) {
  //assert name
  const nameText = findTextByHeading(table, metricAsElement, 'Name');
  expect(nameText).toEqual(metric.name);

  //assert metric template name
  const metricTemplateName = findTextByHeading(
    table,
    metricAsElement,
    'Metric Template',
  );
  expect(metricTemplateName).toEqual(
    metric.metricTemplate ? metric.metricTemplate.name : '-',
  );

  //assert threshold min
  const thresholdMin = findTextByHeading(
    table,
    metricAsElement,
    'Threshold Min',
  );
  expect(thresholdMin).toEqual(
    metric.thresholdRange?.min ? '' + metric.thresholdRange?.min : '-',
  );

  //assert threshold max
  const thresholdMax = findTextByHeading(
    table,
    metricAsElement,
    'Threshold Max',
  );
  expect(thresholdMax).toEqual(
    metric.thresholdRange?.max ? '' + metric.thresholdRange?.max : '-',
  );

  //assert interval
  const intervalText = findTextByHeading(table, metricAsElement, 'Interval');
  expect(intervalText).toEqual(metric.interval);
}

function getCanaryMetric(res: GetCanaryResponse) {
  return _.get(res, ['canary', 'analysis', 'metrics', 0]);
}

function getTableAndRows(selector: string): [Element, NodeListOf<Element>] {
  const tbl = document.querySelector(selector);
  const rows = tbl?.querySelectorAll('tbody tr');
  if (!tbl) {
    throw new Error('Table not found');
  }

  if (!rows) {
    throw new Error('Rows not found');
  }

  return [tbl, rows];
}
