import { act, render, screen } from '@testing-library/react';
import { APIContext } from '../../../contexts/API';
import {
  defaultContexts,
  ProgressiveDeliveryMock,
  withContext,
} from '../../../utils/test-utils';
import CanaryDetails from '../CanaryDetails';

describe('CanaryDetails', () => {
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
          conditions: [
            {
              type: 'Promoted',
              status: 'True',
              lastUpdateTime: '2022-07-18T17:30:00Z',
              lastTransitionTime: 'should be ignored',
              reason: 'Succeeded',
              message: 'some canary status message',
            },
          ],
        },
      },
      automation: {
        kind: 'Kustomization',
        name: 'cool-dep-kustomization',
        namespace: 'some-namespace',
      },
    };

    await act(async () => {
      const c = wrap(
        <CanaryDetails
          name="cool-dep"
          namespace="some-namespace"
          clusterName="my-cluster"
        />,
      );
      render(c);
    });

    // Details
    expect(screen.getByTestId('Cluster')).toHaveTextContent('my-cluster');
    expect(screen.getByTestId('Namespace')).toHaveTextContent('some-namespace');
    expect(screen.getByTestId('Target')).toHaveTextContent(
      'Deployment/cool-dep',
    );
    expect(screen.getByTestId('Application')).toHaveTextContent(
      'Kustomization/cool-dep-kustomization',
    );
    expect(screen.getByTestId('Deployment Strategy')).toHaveTextContent(
      'canary',
    );
    expect(screen.getByTestId('Provider')).toHaveTextContent('my-provider');

    // Status
    expect(screen.getByTestId('phase')).toHaveTextContent('Succeeded');
    expect(screen.getByTestId('failedChecks')).toHaveTextContent('0');
    expect(screen.getByTestId('canaryWeight')).toHaveTextContent('10');
    expect(screen.getByTestId('iterations')).toHaveTextContent('1');
    expect(screen.getByTestId('lastTransitionTime')).toHaveTextContent(
      '2022-07-18T17:25:00Z',
    );

    // Conditions
    expect(screen.getByTestId('type')).toHaveTextContent('Promoted');
    expect(screen.getByTestId('status')).toHaveTextContent('True');
    expect(screen.getByTestId('lastUpdateTime')).toHaveTextContent(
      '2022-07-18T17:30:00Z',
    );
    expect(screen.getByTestId('reason')).toHaveTextContent('Succeeded');
    expect(screen.getByTestId('message')).toHaveTextContent(
      'some canary status message',
    );
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
      const c = wrap(
        <CanaryDetails
          name="cool-dep"
          namespace="some-namespace"
          clusterName="my-cluster"
        />,
      );
      render(c);
    });

    // Details
    expect(screen.getByTestId('Cluster')).toHaveTextContent('my-cluster');
    expect(screen.getByTestId('Namespace')).toHaveTextContent('some-namespace');
    expect(screen.getByTestId('Target')).toHaveTextContent(
      'Deployment/cool-dep',
    );
    expect(screen.getByTestId('Application')).toHaveTextContent('--');
    expect(screen.getByTestId('Deployment Strategy')).toHaveTextContent('--');
    expect(screen.getByTestId('Provider')).toHaveTextContent('--');
  });
});
