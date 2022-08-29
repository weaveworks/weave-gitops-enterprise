import { act, render, screen } from '@testing-library/react';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import {
  defaultContexts,
  PolicyClientMock,
  withContext,
} from '../../../utils/test-utils';
import PolicyDetails from '../PolicyDetails';

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
  it('renders get policy details', async () => {
    api.GetPolicyReturns = {
      policy: {
        name: 'Container Block Sysctls',
        id: 'weave.policies.container-block-sysctl',
        code: 'test policyCode',
        description: 'test description',
        howToSolve: 'test howToSolve',
        category: 'weave.categories.pod-security',
        tags: [
          'pci-dss',
          'cis-benchmark',
          'mitre-attack',
          'nist800-190',
          'gdpr',
          'default',
        ],
        severity: 'high',
        standards: [
          {
            id: 'weave.standards.pci-dss',
            controls: [
              'weave.controls.pci-dss.2.2.4',
              'weave.controls.pci-dss.2.2.5',
            ],
          },
          {
            id: 'weave.standards.cis-benchmark',
            controls: ['weave.controls.cis-benchmark.5.2.6'],
          },
          {
            id: 'weave.standards.mitre-attack',
            controls: ['weave.controls.mitre-attack.4.1'],
          },
          {
            id: 'weave.standards.nist-800-190',
            controls: ['weave.controls.nist-800-190.3.3.1'],
          },
          {
            id: 'weave.standards.gdpr',
            controls: [
              'weave.controls.gdpr.24',
              'weave.controls.gdpr.25',
              'weave.controls.gdpr.32',
            ],
          },
        ],
        gitCommit: '',
        parameters: [
          {
            name: 'exclude_namespace',
            type: 'string',
            value: {
              '@type': 'type.googleapis.com/google.protobuf.StringValue',
              value: 'kube-system',
            },
            required: false,
          },
          {
            name: 'exclude_label_key',
            type: 'string',
            value: null,
            required: false,
          },
          {
            name: 'exclude_label_value',
            type: 'string',
            value: null,
            required: false,
          },
        ],
        targets: {
          kinds: [
            'Deployment',
            'Job',
            'ReplicationController',
            'ReplicaSet',
            'DaemonSet',
            'StatefulSet',
            'CronJob',
          ],
          labels: [],
          namespaces: [],
        },
        createdAt: '2022-08-24T11:11:56Z',
        clusterName: 'default/tw-cluster-2',
        tenant: '',
      },
    };

    await act(async () => {
      const c = wrap(
        <PolicyDetails
          clusterName=""
          id="weave.policies.container-block-sysctl"
        />,
      );
      render(c);
    });

    expect(await screen.findByText('Container Block Sysctls')).toBeTruthy();

    // Details
    expect(screen.getByTestId('Policy ID')).toHaveTextContent(
      'weave.policies.container-block-sysctl',
    );

    expect(screen.queryByTestId('Tenant')).toBeNull();
    
    expect(screen.getByTestId('Cluster Name')).toHaveTextContent(
      'default/tw-cluster-2',
    );

    expect(screen.getByTestId('Severity')).toHaveTextContent('high');
    expect(screen.getByTestId('Category')).toHaveTextContent(
      'weave.categories.pod-security',
    );
    // Tags
    const tags = document.querySelectorAll('#policy-details-header-tags span');
    expect(tags).toHaveLength(6);

     // Targeted K8s Kind
     const kinds = document.querySelectorAll('#policy-details-header-kinds span');
     expect(kinds).toHaveLength(7);

    // description
    expect(screen.getByTestId('description')).toHaveTextContent(
      'test description',
    );

    // how to solve
    expect(screen.getByTestId('howToSolve')).toHaveTextContent(
      'test howToSolve',
    );

    // policyCode
    expect(screen.getByTestId('policyCode')).toHaveTextContent(
      'test policyCode',
    );
  });

  it('renders missing details value', async () => {
    api.GetPolicyReturns = {
      policy: {
        name: 'Container Block Sysctls',
        id: '',
        code: 'test policyCode',
        description: 'test description',
        howToSolve: 'test howToSolve',
        category: '',
        tags: [
         
        ],
        severity: 'high',
        standards: [
          {
            id: 'weave.standards.pci-dss',
            controls: [
              'weave.controls.pci-dss.2.2.4',
              'weave.controls.pci-dss.2.2.5',
            ],
          },
          {
            id: 'weave.standards.cis-benchmark',
            controls: ['weave.controls.cis-benchmark.5.2.6'],
          },
          {
            id: 'weave.standards.mitre-attack',
            controls: ['weave.controls.mitre-attack.4.1'],
          },
          {
            id: 'weave.standards.nist-800-190',
            controls: ['weave.controls.nist-800-190.3.3.1'],
          },
          {
            id: 'weave.standards.gdpr',
            controls: [
              'weave.controls.gdpr.24',
              'weave.controls.gdpr.25',
              'weave.controls.gdpr.32',
            ],
          },
        ],
        gitCommit: '',
        parameters: [
          {
            name: 'exclude_namespace',
            type: 'string',
            value: {
              '@type': 'type.googleapis.com/google.protobuf.StringValue',
              value: 'kube-system',
            },
            required: false,
          },
          {
            name: 'exclude_label_key',
            type: 'string',
            value: null,
            required: false,
          },
          {
            name: 'exclude_label_value',
            type: 'string',
            value: null,
            required: false,
          },
        ],
        targets: {
          kinds: [
           
          ],
          labels: [],
          namespaces: [],
        },
        createdAt: '2022-08-24T11:11:56Z',
        clusterName: '',
        tenant: '',
      },
    };

    await act(async () => {
      const c = wrap(
        <PolicyDetails
          clusterName=""
          id="weave.policies.container-block-sysctl"
        />,
      );
      render(c);
    });

    expect(await screen.findByText('Container Block Sysctls')).toBeTruthy();

    // Details
    expect(screen.getByTestId('Policy ID')).toHaveTextContent(
      '--',
    );
    
    expect(screen.getByTestId('Cluster Name')).toHaveTextContent(
      '--',
    );

    expect(screen.getByTestId('Category')).toHaveTextContent(
      '--',
    );
    // Tags
    expect(screen.getByTestId('Tags')).toHaveTextContent(
      'There is no tags for this policy',
    );
     // Targeted K8s Kind
     expect(screen.getByTestId('Targeted K8s Kind')).toHaveTextContent(
      'There is no kinds for this policy',
    );
    
    // description
    expect(screen.getByTestId('description')).toHaveTextContent(
      'test description',
    );

    // how to solve
    expect(screen.getByTestId('howToSolve')).toHaveTextContent(
      'test howToSolve',
    );

    // policyCode
    expect(screen.getByTestId('policyCode')).toHaveTextContent(
      'test policyCode',
    );
  });
});
