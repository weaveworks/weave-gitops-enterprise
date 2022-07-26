import { act, render, screen } from '@testing-library/react';
import {
  Canary,
  CanaryMetricTemplate,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';

import { ProgressiveDeliveryProvider } from '../../../../contexts/ProgressiveDelivery';
import {
  defaultContexts,
  ProgressiveDeliveryMock,
  withContext,
} from '../../../../utils/test-utils';
import { MetricTemplateModal } from '../Analysis/MetricTemplateModal';

describe('MetricTemplateModal', () => {
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
              "query": "100 - sum(\\n    rate(\\n        istio_requests_total{\\n          reporter=\\"destination\\",\\n          destination_workload_namespace=\\"{{ namespace }}\\",\\n          destination_workload=\\"{{ target }}\\",\\n          response_code!=\\"404\\"\\n        }[{{ interval }}]\\n    )\\n)\\n/\\nsum(\\n    rate(\\n        istio_requests_total{\\n          reporter=\\"destination\\",\\n          destination_workload_namespace=\\"{{ namespace }}\\",\\n          destination_workload=\\"{{ target }}\\"\\n        }[{{ interval }}]\\n    )\\n) * 100\\n",
              "yaml": "apiVersion: flagger.app/v1beta1\\nkind: MetricTemplate\\nmetadata:\\n  annotations:\\n    kubectl.kubernetes.io/last-applied-configuration: |\\n      {\\"apiVersion\\":\\"flagger.app/v1beta1\\",\\"kind\\":\\"MetricTemplate\\",\\"metadata\\":{\\"annotations\\":{},\\"name\\":\\"not-found-percentage\\",\\"namespace\\":\\"canary\\"},\\"spec\\":{\\"provider\\":{\\"address\\":\\"http://prometheus.istio-system:9090\\",\\"type\\":\\"prometheus\\"},\\"query\\":\\"100 - sum(\\\\n    rate(\\\\n        istio_requests_total{\\\\n          reporter=\\\\\\"destination\\\\\\",\\\\n          destination_workload_namespace=\\\\\\"{{ namespace }}\\\\\\",\\\\n          destination_workload=\\\\\\"{{ target }}\\\\\\",\\\\n          response_code!=\\\\\\"404\\\\\\"\\\\n        }[{{ interval }}]\\\\n    )\\\\n)\\\\n/\\\\nsum(\\\\n    rate(\\\\n        istio_requests_total{\\\\n          reporter=\\\\\\"destination\\\\\\",\\\\n          destination_workload_namespace=\\\\\\"{{ namespace }}\\\\\\",\\\\n          destination_workload=\\\\\\"{{ target }}\\\\\\"\\\\n        }[{{ interval }}]\\\\n    )\\\\n) * 100\\\\n\\"}}\\n  creationTimestamp: \\"2022-07-15T16:26:57Z\\"\\n  generation: 1\\n  managedFields:\\n  - apiVersion: flagger.app/v1beta1\\n    fieldsType: FieldsV1\\n    fieldsV1:\\n      f:metadata:\\n        f:annotations:\\n          .: {}\\n          f:kubectl.kubernetes.io/last-applied-configuration: {}\\n      f:spec:\\n        .: {}\\n        f:provider:\\n          .: {}\\n          f:address: {}\\n          f:type: {}\\n        f:query: {}\\n    manager: kubectl-client-side-apply\\n    operation: Update\\n    time: \\"2022-07-15T16:26:57Z\\"\\n  name: not-found-percentage\\n  namespace: canary\\n  resourceVersion: \\"1164\\"\\n  uid: ffa9f533-3cf1-4129-9a6c-3c3913352e8a\\nspec:\\n  provider:\\n    address: http://prometheus.istio-system:9090\\n    type: prometheus\\n  query: |\\n    100 - sum(\\n        rate(\\n            istio_requests_total{\\n              reporter=\\"destination\\",\\n              destination_workload_namespace=\\"{{ namespace }}\\",\\n              destination_workload=\\"{{ target }}\\",\\n              response_code!=\\"404\\"\\n            }[{{ interval }}]\\n        )\\n    )\\n    /\\n    sum(\\n        rate(\\n            istio_requests_total{\\n              reporter=\\"destination\\",\\n              destination_workload_namespace=\\"{{ namespace }}\\",\\n              destination_workload=\\"{{ target }}\\"\\n            }[{{ interval }}]\\n        )\\n    ) * 100\\nstatus: {}\\n"
            }
          }       
        ]
      }
    }`;
    const canary: Canary = JSON.parse(canaryAsJson);
    api.GetCanaryReturns = {
      canary: canary,
    };

    const metrics = canary?.analysis?.metrics;
    const metric = (metrics || [])[0];
    const metricTemplate = metric.metricTemplate as CanaryMetricTemplate;
    const handler = jest.fn();

    await act(async () => {
      const c = wrap(
        <MetricTemplateModal
          open={true}
          metricTemplate={metricTemplate}
          setOpenMetricTemplate={handler}
        />,
      );
      render(c);
    });
    expect(
      await screen.findByText(
        ('Metric Template: ' + metricTemplate.name) as string,
      ),
    ).toBeTruthy();
    ///assert metric template content
    const metricTemplateDialog = document.querySelector(
      '#metric-template-dialog',
    );
    // the expected object is rendered
    expect(metricTemplateDialog?.textContent).toContain('2022-07-15T16:26:57Z');
    // using a consistent yaml view
    expect(metricTemplateDialog?.textContent).toContain(
      'kubectl get metrictemplate not-found-percentage -n canary -o yaml',
    );
  });
});
