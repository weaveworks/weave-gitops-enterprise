import { act, render, screen } from '@testing-library/react';
import { CanaryMetric } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import {
  defaultContexts,
  findTextByHeading,
  ProgressiveDeliveryMock,
  withContext,
} from '../../../../utils/test-utils';
import { CanaryMetricsTable } from '../Analysis/CanaryMetricsTable';
import { APIContext } from '../../../../contexts/API';

describe('CanaryMetricsTable', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: ProgressiveDeliveryMock;

  beforeEach(() => {
    api = new ProgressiveDeliveryMock();
    wrap = withContext([
      ...defaultContexts(),
      [APIContext.Provider, { value: { progressiveDeliveryService: api } }],
    ]);
    api.IsFlaggerAvailableReturns = { clusters: { 'my-cluster': true } };
  });
  it('renders metrics table for a canary with metrics', async () => {
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
    const canary = JSON.parse(canaryAsJson);
    api.GetCanaryReturns = {
      canary: canary,
    };

    await act(async () => {
      const c = wrap(
        <CanaryMetricsTable
          metrics={api.GetCanaryReturns.canary?.analysis?.metrics || []}
        />,
      );
      render(c);
    });

    expect(await screen.findByText('request-success-rate')).toBeTruthy();
    const tbl = document.querySelector('#canary-analysis-metrics table');
    const rows = tbl?.querySelectorAll('tbody tr');
    expect(rows).toHaveLength(1);

    expect(tbl).not.toBeNull();
    expect(rows).not.toBeNull();
    // return early as a type guard
    if (!tbl || !rows) {
      return;
    }

    assertCanaryMetric(
      tbl,
      rows.item(0),
      api.GetCanaryReturns.canary?.analysis?.metrics?.[0] || {},
    );
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
          metrics={api.GetCanaryReturns.canary?.analysis?.metrics || []}
        />,
      );
      render(c);
    });

    expect(await screen.findByText('404s percentage')).toBeTruthy();
    const tbl = document.querySelector('#canary-analysis-metrics table');
    const rows = tbl?.querySelectorAll('tbody tr');
    expect(rows).toHaveLength(1);

    expect(tbl).not.toBeNull();
    expect(rows).not.toBeNull();
    // type guard
    if (!tbl || !rows) {
      return;
    }

    assertCanaryMetric(
      tbl,
      rows.item(0),
      api.GetCanaryReturns.canary?.analysis?.metrics?.[0] || {},
    );
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
