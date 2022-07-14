import { act, render, screen } from '@testing-library/react';
import {CanaryMetricsTable} from '../Analysis/CanaryMetricsTable';
import { ProgressiveDeliveryProvider } from '../../../../contexts/ProgressiveDelivery';
import {
  defaultContexts,
  ProgressiveDeliveryMock,
  withContext,
} from '../../../../utils/test-utils';

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
    var canaryAsJson = `{
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
    "yaml": "interval: 1m\\niterations: 0\\nmirror: false\\nmirrorweight: 0\\nmaxweight: 50\\nstepweight: 10\\nstepweights: []\\nstepweightpromotion: 0\\nthreshold: 5\\nprimaryreadythreshold: null\\ncanaryreadythreshold: null\\nalerts: []\\nmetrics:\\n    - name: request-success-rate\\n      interval: 1m\\n      threshold: 0\\n      thresholdrange:\\n        min: 99\\n        max: null\\n      query: \\"\\"\\n      templateref: null\\n    - name: request-duration\\n      interval: 30s\\n      threshold: 0\\n      thresholdrange:\\n        min: null\\n        max: 500\\n      query: \\"\\"\\n      templateref: null\\nwebhooks:\\n    - type: pre-rollout\\n      name: acceptance-test\\n      url: http://flagger-loadtester.test/\\n      mutealert: false\\n      timeout: 30s\\n      metadata:\\n        cmd: curl -sd 'test' http://canary01-canary:9898/token | grep token\\n        type: bash\\n    - type: \\"\\"\\n      name: load-test\\n      url: http://flagger-loadtester.test/\\n      mutealert: false\\n      timeout: 5s\\n      metadata:\\n        cmd: hey -z 1m -q 10 -c 2 http://canary01-canary.test:9898/\\nmatch: []\\n",
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
  });
});
