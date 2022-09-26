import Applications from '../';

import { MuiThemeProvider } from '@material-ui/core';
import { act, render, RenderResult, screen } from '@testing-library/react';
import {
  CoreClientContextProvider,
  Kind,
  theme,
} from '@weaveworks/weave-gitops';
import { QueryClient, QueryClientProvider } from 'react-query';
import { MemoryRouter } from 'react-router-dom';
import { ThemeProvider } from 'styled-components';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import NotificationsProvider from '../../../contexts/Notifications/Provider';
import RequestContextProvider from '../../../contexts/Request';
import { muiTheme } from '../../../muiTheme';
import {
  CoreClientMock,
  EnterpriseClientMock,
  withContext,
} from '../../../utils/test-utils';

describe('Applications index test', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: CoreClientMock;

  beforeEach(() => {
    api = new CoreClientMock();
    wrap = withContext([
      [ThemeProvider, { theme: theme }],
      [MuiThemeProvider, { theme: muiTheme }],
      [
        RequestContextProvider,
        { fetch: () => new Promise(accept => accept(null)) },
      ],
      [QueryClientProvider, { client: new QueryClient() }],
      [
        EnterpriseClientProvider,
        {
          api: new EnterpriseClientMock(),
        },
      ],
      [CoreClientContextProvider, { api }],
      [MemoryRouter],
      [NotificationsProvider],
    ]);
  });
  it('renders table rows', async () => {
    api.ListObjectsReturns = {
      [Kind.Kustomization]: {
        errors: [],
        objects: [
          {
            uid: 'uid1',
            payload: JSON.stringify({
              // maybe?
              apiVersion: 'kustomize.toolkit.fluxcd.io/v1beta2',
              kind: 'Kustomization',
              metadata: {
                namespace: 'my-ns',
                name: 'my-kustomization',
                uid: 'uid1',
              },
              spec: {
                path: './',
                interval: {},
                sourceRef: {},
              },
              status: {
                conditions: [],
                lastAppliedRevision: '',
                lastAttemptedRevision: '',
                inventory: [],
              },
            }),
            clusterName: 'my-cluster',
          },
        ],
      },
    };

    await act(async () => {
      const c = wrap(<Applications />);
      render(c);
    });

    expect(await screen.findByText('my-kustomization')).toBeTruthy();
  });

  describe('snapshots', () => {
    it('loading', async () => {
      await act(async () => {
        const c = wrap(<Applications />);
        const result = render(c);

        expect(result.container).toMatchSnapshot();
      });
    });
    it('success', async () => {
      let result: RenderResult;
      await act(async () => {
        const c = wrap(<Applications />);
        result = await render(c);
      });

      //   @ts-ignore
      expect(result.container).toMatchSnapshot();
    });
  });
});

// const foo = {
//   apiVersion: 'kustomize.toolkit.fluxcd.io/v1beta2',
//   kind: 'Kustomization',
//   metadata: {
//     name: 'canaries',
//     namespace: 'flux-system',
//     resourceVersion: '158728446',
//     uid: 'ebcafd35-ef1f-4dcf-b386-b615c141b955',
//   },
//   spec: {
//     force: false,
//     interval: '10m0s',
//     path: './canaries',
//     prune: true,
//     sourceRef: {
//       kind: 'GitRepository',
//       name: 'flux-system',
//     },
//   },
//   status: {
//     conditions: [
//       {
//         lastTransitionTime: '2022-09-26T16:28:12Z',
//         message:
//           'Deployment/blue-green/podinfo dry-run failed, reason: ==================================================================\nPolicy\t: weave.policies.controller-serviceaccount-tokens-automount\nEntity\t: deployment/podinfo in namespace: blue-green\nOccurrences:\n- \'automountServiceAccountToken\' must be set; found \'{"containers": [{"command": ["./podinfo", "--port=9898", "--port-metrics=9797", "--grpc-port=9999", "--grpc-service-name=podinfo", "--level=info", "--random-delay=false", "--random-error=false"], "env": [{"name": "PODINFO_UI_COLOR", "value": "#34577c"}], "image": "ghcr.io/stefanprodan/podinfo:6.0.1", "imagePullPolicy": "IfNotPresent", "livenessProbe": {"exec": {"command": ["podcli", "check", "http", "localhost:9898/healthz"]}, "failureThreshold": 3, "initialDelaySeconds": 5, "periodSeconds": 10, "successThreshold": 1, "timeoutSeconds": 5}, "name": "podinfod", "ports": [{"containerPort": 9898, "name": "http", "protocol": "TCP"}, {"containerPort": 9797, "name": "http-metrics", "protocol": "TCP"}, {"containerPort": 9999, "name": "grpc", "protocol": "TCP"}], "readinessProbe": {"exec": {"command": ["podcli", "check", "http", "localhost:9898/readyz"]}, "failureThreshold": 3, "initialDelaySeconds": 5, "periodSeconds": 10, "successThreshold": 1, "timeoutSeconds": 5}, "resources": {"limits": {"cpu": "2", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "64Mi"}}, "terminationMessagePath": "/dev/termination-log", "terminationMessagePolicy": "File"}], "dnsPolicy": "ClusterFirst", "restartPolicy": "Always", "schedulerName": "default-scheduler", "securityContext": {}, "terminationGracePeriodSeconds": 30}\'\n, error: admission webhook "admission.agent.weaveworks" denied the request: ==================================================================\nPolicy\t: weave.policies.controller-serviceaccount-tokens-automount\nEntity\t: deployment/podinfo in namespace: blue-green\nOccurrences:\n- \'automountServiceAccountToken\' must be set; found \'{"containers": [{"command": ["./podinfo", "--port=9898", "--port-metrics=9797", "--grpc-port=9999", "--grpc-service-name=podinfo", "--level=info", "--random-delay=false", "--random-error=false"], "env": [{"name": "PODINFO_UI_COLOR", "value": "#34577c"}], "image": "ghcr.io/stefanprodan/podinfo:6.0.1", "imagePullPolicy": "IfNotPresent", "livenessProbe": {"exec": {"command": ["podcli", "check", "http", "localhost:9898/healthz"]}, "failureThreshold": 3, "initialDelaySeconds": 5, "periodSeconds": 10, "successThreshold": 1, "timeoutSeconds": 5}, "name": "podinfod", "ports": [{"containerPort": 9898, "name": "http", "protocol": "TCP"}, {"containerPort": 9797, "name": "http-metrics", "protocol": "TCP"}, {"containerPort": 9999, "name": "grpc", "protocol": "TCP"}], "readinessProbe": {"exec": {"command": ["podcli", "check", "http", "localhost:9898/readyz"]}, "failureThreshold": 3, "initialDelaySeconds": 5, "periodSeconds": 10, "successThreshold": 1, "timeoutSeconds": 5}, "resources": {"limits": {"cpu": "2", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "64Mi"}}, "terminationMessagePath": "/dev/termination-log", "terminationMessagePolicy": "File"}], "dnsPolicy": "ClusterFirst", "restartPolicy": "Always", "schedulerName": "default-scheduler", "securityContext": {}, "terminationGracePeriodSeconds": 30}\'\n\n',
//         reason: 'ReconciliationFailed',
//         status: 'False',
//         type: 'Ready',
//       },
//     ],
//     lastAttemptedRevision: 'main/ac4fe2a533a8a3ce0b88d43c995b69500db43f0a',
//     observedGeneration: 1,
//   },
// };
