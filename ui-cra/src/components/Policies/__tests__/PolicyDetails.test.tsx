import { act, render, screen } from '@testing-library/react';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import {
  defaultContexts,
  PolicyClientMock,
  withContext,
} from '../../../utils/test-utils';
import PolicyDetails from '../PolicyDetails';
import { parseValue } from '../PolicyDetails/PolicyUtilis';
const MockPolicyResponse = {
  policy: {
    name: 'Containers Sharing Host PID',
    id: 'weave.policies.containers-sharing-host-pid',
    code: '1package weave.advisor.podSecurity.deny_shared_host_pid',
    description:
      'This Policy allows check if sharing host PID namespace with the container should be allowed or not.',
    howToSolve:
      'Match the shared resource with either true or false, as set in your constraint.',
    category: 'weave.categories.pod-security',
    tags: ['cis-benchmark', 'nist800-190', 'gdpr', 'default'],
    severity: 'high',
    standards: [],
    gitCommit: '',
    parameters: [
      {
        name: 'resource_enabled',
        type: 'boolean',
        value: {
          '@type': 'type.googleapis.com/google.protobuf.BoolValue',
          value: false,
        },
        required: true,
      },
      {
        name: 'exclude_namespace',
        type: 'string',
        value: null,
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
    createdAt: '2022-08-31T16:46:14Z',
    clusterName: 'management',
    tenant: '',
    modes: ['audit', 'admission'],
  },
};
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
    const policy = MockPolicyResponse.policy;
    api.GetPolicyReturns = MockPolicyResponse;

    await act(async () => {
      const c = wrap(<PolicyDetails clusterName="" id={policy.id} />);
      render(c);
    });

    expect(await screen.findByText(policy.name)).toBeTruthy();

    // Details
    expect(screen.getByTestId('Policy ID')).toHaveTextContent(policy.id);

    expect(screen.queryByTestId('Tenant')).toHaveTextContent(
      policy.tenant || '--',
    );

    expect(screen.getByTestId('Cluster')).toHaveTextContent(policy.clusterName);

    expect(screen.getByTestId('Severity')).toHaveTextContent(policy.severity);
    expect(screen.getByTestId('Category')).toHaveTextContent(policy.category);
    // Tags
    const tags = document.querySelectorAll('#policy-details-header-tags span');
    expect(tags).toHaveLength(policy.tags.length);

    // Targeted K8s Kind
    const kinds = document.querySelectorAll(
      '#policy-details-header-kinds span',
    );
    expect(kinds).toHaveLength(policy.targets.kinds.length);

    // description
    const desc = document.querySelector(
      'div[data-testid="description"] > .editor',
    );
    expect(desc).toHaveTextContent(policy.description);

    // how to solve
    expect(
      document.querySelector('div[data-testid="howToSolve"] > .editor'),
    ).toHaveTextContent(policy.howToSolve);

    // policyCode
    expect(
      document.querySelector('div[data-testid="policyCode"] code'),
    ).toHaveTextContent(policy.code);

    // Parameters
    policy.parameters.forEach(parameter => {
      const paramWrapper = document.getElementById(parameter.name);
      // Name
      ValidateParameter(paramWrapper, 'Name', parameter.name);
      // Type
      ValidateParameter(paramWrapper, 'Type', parameter.type);
      // Value
      ValidateParameter(paramWrapper, 'Value', parseValue(parameter));
      // Required
      ValidateParameter(
        paramWrapper,
        'Required',
        parameter.required ? 'True' : 'False',
      );
    });
  });

  it('renders missing details value', async () => {
    const policy = MockPolicyResponse.policy;
    policy.tags = [];
    policy.severity = '';
    policy.category = '';
    policy.targets.kinds = [];
    api.GetPolicyReturns = { policy };

    await act(async () => {
      const c = wrap(<PolicyDetails clusterName="" id={policy.id} />);
      render(c);
    });

    expect(await screen.findByText(policy.name)).toBeTruthy();

    // Details
    expect(screen.getByTestId('Severity')).toHaveTextContent('Severity :');

    expect(screen.getByTestId('Category')).toHaveTextContent('--');
    // Tags
    expect(screen.getByTestId('Tags')).toHaveTextContent(
      'There is no tags for this policy',
    );
    // Targeted K8s Kind
    expect(screen.getByTestId('Targeted K8s Kind')).toHaveTextContent(
      'There is no kinds for this policy',
    );
  });
});

const ValidateParameter = (
  paramElement: HTMLElement | null,
  key: string,
  value: string | unknown,
) => {
  const parameterNameWrapper = paramElement?.querySelector(
    `div[data-testid="${key}"]`,
  );
  const label = parameterNameWrapper?.querySelector('.label');
  const body1 = parameterNameWrapper?.querySelector('.body1');
  expect(label?.textContent).toEqual(key);

  // Handel ChipWrapper of the Parsing Value
  if (body1?.textContent === 'undefined') {
    expect(body1?.textContent).toEqual(`${(value as Element)?.textContent}`);
  } else {
    expect(body1?.textContent).toEqual(value);
  }
};
