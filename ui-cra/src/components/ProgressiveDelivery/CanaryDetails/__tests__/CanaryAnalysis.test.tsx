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
  "targetReference": {
    "kind": "Deployment",
    "name": "deployment01"
  },
  "targetDeployment": {
    "uid": "37b1b1b5-a8e4-40c0-94bb-8ab50840dd76",
    "resourceVersion": "1841",
    "fluxLabels": {
    },
    "appliedImageVersions": {
      "app": "ghcr.io/yitsushi/hello-world:1.0.7"
    },
    "promotedImageVersions": {
      "app": "ghcr.io/yitsushi/hello-world:1.0.7"
    }
  },
  "status": {
    "phase": "Initializing",
    "lastTransitionTime": "2022-07-11T08:04:51Z",
    "conditions": [
      {
        "type": "Promoted",
        "status": "Unknown",
        "lastUpdateTime": "2022-07-11T08:04:51Z",
        "lastTransitionTime": "2022-07-11T08:04:51Z",
        "reason": "Initializing",
        "message": "New Deployment detected, starting initialization."
      }
    ]
  },
  "deploymentStrategy": "canary",
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
  console.log(table.textContent)

  //assert name
  const nameText = findTextByHeading(table, metricAsElement, 'Name')
  expect(nameText).toEqual(metric.name);

  //assert namespace
  const namespaceText = findTextByHeading(table, metricAsElement, 'Ns')
  expect(namespaceText).toEqual(metric.namespace || "-");

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
    if (e.innerHTML.includes(heading)) {
      idx = i;
    }
  });

  return idx;
}

