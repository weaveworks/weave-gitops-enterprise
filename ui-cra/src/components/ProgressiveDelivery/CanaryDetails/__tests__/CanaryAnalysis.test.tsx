import { act, render, screen } from '@testing-library/react';
import {CanaryMetricsTable} from '../Analysis/CanaryMetricsTable';
import { ProgressiveDeliveryProvider } from '../../../../contexts/ProgressiveDelivery';
import {
  defaultContexts,
  ProgressiveDeliveryMock,
  withContext,
} from '../../../../utils/test-utils';
import {CanaryMetric} from "@weaveworks/progressive-delivery/api/prog/types.pb";


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
    let canaryAsJson = `{
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
          "min": 99
        },
        "interval": "1m"
      },
      {
        "name": "request-duration",
        "thresholdRange": {
          "max": 500
        },
        "interval": "30s"
      }
    ]
  }
}`
    let canary = JSON.parse(canaryAsJson)
    api.GetCanaryReturns = {
      canary: canary
    };

    await act(async () => {
      const c = wrap(<CanaryMetricsTable  metrics={api.GetCanaryReturns.canary?.analysis?.metrics}/>);
      render(c);
    });

    expect(await screen.findByText('request-success-rate')).toBeTruthy();
    const tbl = document.querySelector('#canary-analysis-metrics table');
    const rows = tbl?.querySelectorAll('tbody tr');
    expect(rows).toHaveLength(2);
    assertCanaryMetric(tbl,rows.item(0),api.GetCanaryReturns.canary?.analysis?.metrics[0])
    assertCanaryMetric(tbl,rows.item(1),api.GetCanaryReturns.canary?.analysis?.metrics[1])
  });
});

function assertCanaryMetric(table: Element, metricAsElement: Element, metric: CanaryMetric) {
  //assert name
  const nameText = findTextByHeading(table, metricAsElement, 'Name')
  expect(nameText).toEqual(metric.name);

  //assert namespace
  const namespaceText = findTextByHeading(table, metricAsElement, 'Namespace')
  expect(namespaceText).toEqual(metric.namespace ? metric.namespace : "-");

  //assert threshold min
  const thresholdMin = findTextByHeading(table, metricAsElement, 'Threshold Min')
  expect(thresholdMin).toEqual(metric.thresholdRange?.min ? ""+metric.thresholdRange?.min : "-");

  //assert threshold max
  const thresholdMax = findTextByHeading(table, metricAsElement, 'Threshold Max')
  expect(thresholdMax).toEqual(metric.thresholdRange?.max ? ""+metric.thresholdRange?.max : "-");

  //assert interval
  const intervalText = findTextByHeading(table, metricAsElement, 'Interval')
  expect(intervalText).toEqual(metric.interval);
}

function findTextByHeading(table: Element,row: Element, headingName: string) {
  const cols = table?.querySelectorAll('thead th');
  const index = findColByHeading(cols, headingName) as number;
  return row.childNodes.item(index).textContent;
}


// Helper to ensure that tests still pass if columns get re-ordered
function findColByHeading(
    cols: NodeListOf<Element> | undefined,
    heading: string,
): null | number {
  if (!cols) {
    return null;
  }

  let idx = null;
  cols?.forEach((e, i) => {
    //TODO: look for a better matching
    if (e.innerHTML.match("(>"+heading+"<)")) {
      idx = i;
    }
  });

  return idx;
}

