import { act, render, screen } from '@testing-library/react';
import CanaryDetails from '../CanaryDetails';
import { ProgressiveDeliveryProvider } from '../../../contexts/ProgressiveDelivery';
import {
  defaultContexts,
  ProgressiveDeliveryMock,
  withContext,
} from '../../../utils/test-utils';

describe('CanaryDetails', () => {
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

  it('renders canary details', async () => {
    api.GetCanaryReturns = {
      canary: {
        name: 'my-canary',
        namespace: 'some-namespace',
        clusterName: 'my-cluster',
        targetReference: {
          kind: 'Deployment',
          name: 'cool-dep',
        },
        deploymentStrategy: 'canary',
        provider: 'my-provider',
        status: {
          phase: 'Succeeded',
          failedChecks: 0,
          canaryWeight: 10,
          iterations: 1,
          lastTransitionTime: '2022-07-18T17:25:00Z',
          conditions: [{
            type: 'Promoted',
            status: 'True',
            lastUpdateTime: '2022-07-18T17:30:00Z',
            lastTransitionTime: 'should be ignored',
            reason: 'Succeeded',
            message: 'some canary status message'
          }]
        }
      },
      automation: {
        kind: 'Kustomization',
        name: 'cool-dep-kustomization',
        namespace: 'some-namespace'
      },
    };

    await act(async () => {
      const c = wrap(<CanaryDetails name='cool-dep' namespace='some-namespace' clusterName='my-cluster' />);
      render(c);
    });

    // Details
    expect(await screen.getByTestId('Cluster')).toHaveTextContent('my-cluster');
    expect(await screen.getByTestId('Namespace')).toHaveTextContent('some-namespace');
    expect(await screen.getByTestId('Target')).toHaveTextContent('Deployment/cool-dep');
    expect(await screen.getByTestId('Application')).toHaveTextContent('Kustomization/cool-dep-kustomization');
    expect(await screen.getByTestId('Deployment Strategy')).toHaveTextContent('canary');
    expect(await screen.getByTestId('Provider')).toHaveTextContent('my-provider');

    // Status
    expect(await screen.getByTestId('phase')).toHaveTextContent('Succeeded');
    expect(await screen.getByTestId('failedChecks')).toHaveTextContent('0');
    expect(await screen.getByTestId('canaryWeight')).toHaveTextContent('10');
    expect(await screen.getByTestId('iterations')).toHaveTextContent('1');
    expect(await screen.getByTestId('lastTransitionTime')).toHaveTextContent('2022-07-18T17:25:00Z');

    // Conditions
    expect(await screen.getByTestId('type')).toHaveTextContent('Promoted');
    expect(await screen.getByTestId('status')).toHaveTextContent('True');
    expect(await screen.getByTestId('lastUpdateTime')).toHaveTextContent('2022-07-18T17:30:00Z');
    expect(await screen.getByTestId('reason')).toHaveTextContent('Succeeded');
    expect(await screen.getByTestId('message')).toHaveTextContent('some canary status message');
  });

  it('renders canary with missing optional details', async () => {
    api.GetCanaryReturns = {
      canary: {
        name: 'my-canary',
        namespace: 'some-namespace',
        clusterName: 'my-cluster',
        targetReference: {
          kind: 'Deployment',
          name: 'cool-dep',
        },
      },
    };

    await act(async () => {
      const c = wrap(<CanaryDetails name='cool-dep' namespace='some-namespace' clusterName='my-cluster' />);
      render(c);
    });

    // Details
    expect(await screen.getByTestId('Cluster')).toHaveTextContent('my-cluster');
    expect(await screen.getByTestId('Namespace')).toHaveTextContent('some-namespace');
    expect(await screen.getByTestId('Target')).toHaveTextContent('Deployment/cool-dep');
    expect(await screen.getByTestId('Application')).toHaveTextContent('--');
    expect(await screen.getByTestId('Deployment Strategy')).toHaveTextContent('--');
    expect(await screen.getByTestId('Provider')).toHaveTextContent('--');
  });
});
