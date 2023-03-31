import { act, render, screen } from '@testing-library/react';
import moment from 'moment';
import { describe, expect, it } from 'vitest';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import {
  defaultContexts,
  PolicyClientMock,
  withContext,
} from '../../../utils/test-utils';
import PolicyViolationDetails from '../ViolationDetails';

describe('ListPolicViolations', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: PolicyClientMock;

  beforeEach(() => {
    api = new PolicyClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [EnterpriseClientProvider, { api }],
    ]);
  });
  it('renders policy violation details', async () => {
    api.GetPolicyValidationReturns = {
      violation: {
        id: '2c1d87a4-525c-4c54-a587-c6f6a904ca31',
        message:
          'Controller ServiceAccount Tokens Automount in deployment helm-controller (1 occurrences)',
        clusterId: '659dc1ec-35b4-4d1d-a1de-9371cefcf81e',
        category: 'weave.categories.access-control',
        severity: 'high',
        createdAt: '2022-08-24T19:39:11Z',
        entity: 'helm-controller',
        namespace: 'flux-system',
        violatingEntity: 'test violatingEntity',
        description: 'test description',
        howToSolve: 'test howToSolve',
        name: 'Controller ServiceAccount Tokens Automount',
        clusterName: 'default/tw-cluster-2',
        occurrences: [
          {
            message:
              '\'automountServiceAccountToken\' must be set; found \'{"containers": [{"args": ["--events-addr=http://notification-controller.flux-system.svc.cluster.local./", "--watch-all-namespaces=true", "--log-level=info", "--log-encoding=json", "--enable-leader-election"], "env": [{"name": "RUNTIME_NAMESPACE", "valueFrom": {"fieldRef": {"apiVersion": "v1", "fieldPath": "metadata.namespace"}}}], "image": "ghcr.io/fluxcd/helm-controller:v0.22.2", "imagePullPolicy": "IfNotPresent", "livenessProbe": {"failureThreshold": 3, "httpGet": {"path": "/healthz", "port": "healthz", "scheme": "HTTP"}, "periodSeconds": 10, "successThreshold": 1, "timeoutSeconds": 1}, "name": "manager", "ports": [{"containerPort": 8080, "name": "http-prom", "protocol": "TCP"}, {"containerPort": 9440, "name": "healthz", "protocol": "TCP"}], "readinessProbe": {"failureThreshold": 3, "httpGet": {"path": "/readyz", "port": "healthz", "scheme": "HTTP"}, "periodSeconds": 10, "successThreshold": 1, "timeoutSeconds": 1}, "resources": {"limits": {"cpu": "1", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "64Mi"}}, "securityContext": {"allowPrivilegeEscalation": false, "capabilities": {"drop": ["ALL"]}, "readOnlyRootFilesystem": true, "runAsNonRoot": true, "seccompProfile": {"type": "RuntimeDefault"}}, "terminationMessagePath": "/dev/termination-log", "terminationMessagePolicy": "File", "volumeMounts": [{"mountPath": "/tmp", "name": "temp"}]}], "dnsPolicy": "ClusterFirst", "nodeSelector": {"kubernetes.io/os": "linux"}, "restartPolicy": "Always", "schedulerName": "default-scheduler", "securityContext": {"fsGroup": 1337}, "serviceAccount": "helm-controller", "serviceAccountName": "helm-controller", "terminationGracePeriodSeconds": 600, "volumes": [{"emptyDir": {}, "name": "temp"}]}\'',
          },
        ],
      },
    };

    await act(async () => {
      const c = wrap(
        <PolicyViolationDetails
          clusterName="default/tw-cluster-2"
          id="2c1d87a4-525c-4c54-a587-c6f6a904ca31"
        />,
      );
      render(c);
    });
    //Violation Logs
    expect(await screen.findByText('Violation Logs')).toBeTruthy();

    // Details

    expect(screen.getByTestId('Cluster')).toHaveTextContent(
      'default/tw-cluster-2',
    );
    expect(screen.getByTestId('Violation Time')).toHaveTextContent(
      moment('2022-08-24T19:39:11Z').fromNow(),
    );
    expect(screen.getByTestId('Severity')).toHaveTextContent('high');
    expect(screen.getByTestId('Category')).toHaveTextContent(
      'weave.categories.access-control',
    );
    expect(screen.getByTestId('Application')).toHaveTextContent(
      'flux-system/helm-controller',
    );

    // Occurrences
    const occurrences = document.querySelectorAll('#occurrences li');
    expect(occurrences).toHaveLength(1);

    // description
    expect(screen.getByTestId('description')).toHaveTextContent(
      'test description',
    );

    // how to solve
    expect(screen.getByTestId('howToSolve')).toHaveTextContent(
      'test howToSolve',
    );

    // Violating Entity
    expect(screen.getByTestId('violatingEntity')).toHaveTextContent(
      'test violatingEntity',
    );
  });

  it('renders policy missing violation details', async () => {
    api.GetPolicyValidationReturns = {
      violation: {
        id: '2c1d87a4-525c-4c54-a587-c6f6a904ca31',
        message:
          'Controller ServiceAccount Tokens Automount in deployment helm-controller (1 occurrences)',
        clusterId: '659dc1ec-35b4-4d1d-a1de-9371cefcf81e',
        category: '',
        severity: 'test',
        createdAt: '',
        entity: 'helm-controller',
        namespace: 'flux-system',
        violatingEntity: 'test violatingEntity',
        description: 'test description',
        howToSolve: 'test howToSolve',
        name: 'Controller ServiceAccount Tokens Automount',
        clusterName: '',
        occurrences: [
          {
            message:
              '\'automountServiceAccountToken\' must be set; found \'{"containers": [{"args": ["--events-addr=http://notification-controller.flux-system.svc.cluster.local./", "--watch-all-namespaces=true", "--log-level=info", "--log-encoding=json", "--enable-leader-election"], "env": [{"name": "RUNTIME_NAMESPACE", "valueFrom": {"fieldRef": {"apiVersion": "v1", "fieldPath": "metadata.namespace"}}}], "image": "ghcr.io/fluxcd/helm-controller:v0.22.2", "imagePullPolicy": "IfNotPresent", "livenessProbe": {"failureThreshold": 3, "httpGet": {"path": "/healthz", "port": "healthz", "scheme": "HTTP"}, "periodSeconds": 10, "successThreshold": 1, "timeoutSeconds": 1}, "name": "manager", "ports": [{"containerPort": 8080, "name": "http-prom", "protocol": "TCP"}, {"containerPort": 9440, "name": "healthz", "protocol": "TCP"}], "readinessProbe": {"failureThreshold": 3, "httpGet": {"path": "/readyz", "port": "healthz", "scheme": "HTTP"}, "periodSeconds": 10, "successThreshold": 1, "timeoutSeconds": 1}, "resources": {"limits": {"cpu": "1", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "64Mi"}}, "securityContext": {"allowPrivilegeEscalation": false, "capabilities": {"drop": ["ALL"]}, "readOnlyRootFilesystem": true, "runAsNonRoot": true, "seccompProfile": {"type": "RuntimeDefault"}}, "terminationMessagePath": "/dev/termination-log", "terminationMessagePolicy": "File", "volumeMounts": [{"mountPath": "/tmp", "name": "temp"}]}], "dnsPolicy": "ClusterFirst", "nodeSelector": {"kubernetes.io/os": "linux"}, "restartPolicy": "Always", "schedulerName": "default-scheduler", "securityContext": {"fsGroup": 1337}, "serviceAccount": "helm-controller", "serviceAccountName": "helm-controller", "terminationGracePeriodSeconds": 600, "volumes": [{"emptyDir": {}, "name": "temp"}]}\'',
          },
        ],
      },
    };

    await act(async () => {
      const c = wrap(
        <PolicyViolationDetails
          clusterName="default/tw-cluster-2"
          id="2c1d87a4-525c-4c54-a587-c6f6a904ca31"
        />,
      );
      render(c);
    });

    expect(screen.getByTestId('Cluster')).toHaveTextContent('--');
    expect(screen.getByTestId('Severity')).toHaveTextContent('test');
    expect(screen.getByTestId('Category')).toHaveTextContent('--');
  });

  it('renders application violation details', async () => {
    api.GetPolicyValidationReturns = {
      violation: {
        id: '2c1d87a4-525c-4c54-a587-c6f6a904ca31',
        message:
          'Controller ServiceAccount Tokens Automount in deployment helm-controller (1 occurrences)',
        clusterId: '659dc1ec-35b4-4d1d-a1de-9371cefcf81e',
        category: 'weave.categories.access-control',
        severity: 'high',
        createdAt: '2022-08-24T19:39:11Z',
        entity: 'helm-controller',
        namespace: 'flux-system',
        violatingEntity: 'test violatingEntity',
        description: 'test description',
        howToSolve: 'test howToSolve',
        name: 'Controller ServiceAccount Tokens Automount',
        clusterName: 'default/tw-cluster-2',
        occurrences: [
          {
            message:
              '\'automountServiceAccountToken\' must be set; found \'{"containers": [{"args": ["--events-addr=http://notification-controller.flux-system.svc.cluster.local./", "--watch-all-namespaces=true", "--log-level=info", "--log-encoding=json", "--enable-leader-election"], "env": [{"name": "RUNTIME_NAMESPACE", "valueFrom": {"fieldRef": {"apiVersion": "v1", "fieldPath": "metadata.namespace"}}}], "image": "ghcr.io/fluxcd/helm-controller:v0.22.2", "imagePullPolicy": "IfNotPresent", "livenessProbe": {"failureThreshold": 3, "httpGet": {"path": "/healthz", "port": "healthz", "scheme": "HTTP"}, "periodSeconds": 10, "successThreshold": 1, "timeoutSeconds": 1}, "name": "manager", "ports": [{"containerPort": 8080, "name": "http-prom", "protocol": "TCP"}, {"containerPort": 9440, "name": "healthz", "protocol": "TCP"}], "readinessProbe": {"failureThreshold": 3, "httpGet": {"path": "/readyz", "port": "healthz", "scheme": "HTTP"}, "periodSeconds": 10, "successThreshold": 1, "timeoutSeconds": 1}, "resources": {"limits": {"cpu": "1", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "64Mi"}}, "securityContext": {"allowPrivilegeEscalation": false, "capabilities": {"drop": ["ALL"]}, "readOnlyRootFilesystem": true, "runAsNonRoot": true, "seccompProfile": {"type": "RuntimeDefault"}}, "terminationMessagePath": "/dev/termination-log", "terminationMessagePolicy": "File", "volumeMounts": [{"mountPath": "/tmp", "name": "temp"}]}], "dnsPolicy": "ClusterFirst", "nodeSelector": {"kubernetes.io/os": "linux"}, "restartPolicy": "Always", "schedulerName": "default-scheduler", "securityContext": {"fsGroup": 1337}, "serviceAccount": "helm-controller", "serviceAccountName": "helm-controller", "terminationGracePeriodSeconds": 600, "volumes": [{"emptyDir": {}, "name": "temp"}]}\'',
          },
        ],
      },
    };

    await act(async () => {
      const c = wrap(
        <PolicyViolationDetails
          clusterName="default/tw-cluster-2"
          id="2c1d87a4-525c-4c54-a587-c6f6a904ca31"
          source="APPLICATION"
          sourcePath="kustomization"
        />,
      );
      render(c);
    });
    //Violation Logs
    expect(await screen.findByText('helm-controller')).toBeTruthy();

    // Details
    expect(screen.getByTestId('Cluster')).toHaveTextContent(
      'default/tw-cluster-2',
    );

    expect(screen.getByTestId('Policy Name')).toHaveTextContent(
      'Controller ServiceAccount Tokens Automount',
    );

    expect(screen.getByTestId('Violation Time')).toHaveTextContent(
      moment('2022-08-24T19:39:11Z').fromNow(),
    );
    expect(screen.getByTestId('Severity')).toHaveTextContent('high');
    expect(screen.getByTestId('Category')).toHaveTextContent(
      'weave.categories.access-control',
    );
    expect(screen.queryByTestId('Application')).toBeNull();
  });
});
