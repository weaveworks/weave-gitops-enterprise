import { act, render, screen } from '@testing-library/react';
import moment from 'moment';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import {
  defaultContexts,
  PolicyConfigsClientMock,
  withContext,
} from '../../../utils/test-utils';
import PolicyConfigDetails from '../PolicyConfigDetails';
import { renderParameterValue } from '../PolicyConfigDetails/PolicyDetailsCard';

const MockPolicyConfigDetailsResponse = {
  policyConfig: {
    name: 'policyconfig name',
    clusterName: 'management',
    age: '2022-08-31T16:46:14Z',
    status: 'Warning',
    matchType: 'namespaces',
    match: {
      namespaces: ['namespace one', 'namespace two'],
      workspaces: ['workspace one', 'workspace two'],
      apps: [
        {
          name: 'app name',
          kind: 'app kind',
          namespace: 'app namespace',
        },
        {
          name: 'app name two',
          kind: 'app kind two',
          namespace: '',
        },
      ],
      resources: [
        {
          name: 'resource name',
          kind: 'resource kind',
          namespace: 'resource namespace',
        },
      ],
    },
    policies: [
      {
        id: 'policy-id-number-one',
        name: 'policy number one',
        description: '',
        parameters: { replica_count: 3, run_as_route: true },
        status: 'OK',
      },
      {
        id: 'policy-id-number-two',
        name: 'policy number two',
        description: '',
        parameters: { replica_count: 3, run_as_route: true },
        status: 'Warning',
      },
    ],
    totalPolicies: 2,
  },
};

describe('GetPolicyConfigDetails', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: PolicyConfigsClientMock;

  beforeEach(() => {
    api = new PolicyConfigsClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [EnterpriseClientProvider, { api }],
    ]);
  });
  it('renders get policyConfig details with Applied Namespaces', async () => {
    const policyConfig = MockPolicyConfigDetailsResponse.policyConfig;
    api.GetPolicyConfigReturns = policyConfig;
    const matchedItem = policyConfig.match.namespaces;

    await act(async () => {
      const c = wrap(
        <PolicyConfigDetails
          clusterName={policyConfig.clusterName}
          name={policyConfig.name}
        />,
      );
      render(c);
    });

    expect(await screen.findByText(policyConfig.name)).toBeTruthy();

    //Header Section details

    expect(screen.getByTestId('Cluster')).toHaveTextContent(
      policyConfig.clusterName,
    );

    expect(screen.getByTestId('Age')).toHaveTextContent(
      moment(policyConfig.age).fromNow(),
    );
    const AppliedTo = document.querySelector(
      'span[data-testid="appliedTo"]',
    );
    expect(AppliedTo).toHaveTextContent(`${policyConfig.matchType} (${matchedItem.length})`);

    matchedItem.map(item => {
      const AppliedToItem = document.querySelector(
        `li[data-testid="matchItem${item}"]`,
      );
      expect(AppliedToItem).toHaveTextContent(item);
    });
  });

  it('renders applied Apps', async () => {
    const policyConfig = MockPolicyConfigDetailsResponse.policyConfig;
    const matchedItem = policyConfig.match.apps;
    api.GetPolicyConfigReturns = policyConfig;

    policyConfig.matchType = 'apps';
    await act(async () => {
      const c = wrap(
        <PolicyConfigDetails
          clusterName={policyConfig.clusterName}
          name={policyConfig.name}
        />,
      );
      render(c);
    });

    matchedItem.map(item => {
      const AppliedToItem = document.querySelector(
        `span[data-testid="matchItem${item.name}"]`,
      );
      const AppliedToKind = document.querySelector(
        `span[data-testid="matchItemKind${item.kind}"]`,
      );
      expect(AppliedToItem).toHaveTextContent(
        `${item.namespace || '*'}/${item.name}`,
      );
      expect(AppliedToKind).toHaveTextContent(item.kind);
    });
  });

  it('renders applied Resources', async () => {
    const policyConfig = MockPolicyConfigDetailsResponse.policyConfig;
    const matchedItem = policyConfig.match.resources;
    api.GetPolicyConfigReturns = policyConfig;

    policyConfig.matchType = 'resources';
    await act(async () => {
      const c = wrap(
        <PolicyConfigDetails
          clusterName={policyConfig.clusterName}
          name={policyConfig.name}
        />,
      );
      render(c);
    });

    matchedItem.map(item => {
      const AppliedToItem = document.querySelector(
        `span[data-testid="matchItem${item.name}"]`,
      );
      const AppliedToKind = document.querySelector(
        `span[data-testid="matchItemKind${item.kind}"]`,
      );
      expect(AppliedToItem).toHaveTextContent(
        `${item.namespace || '*'}/${item.name}`,
      );
      expect(AppliedToKind).toHaveTextContent(item.kind);
    });
  });

  it('renders applied Workspaces', async () => {
    const policyConfig = MockPolicyConfigDetailsResponse.policyConfig;
    const matchedItem = policyConfig.match.workspaces;
    api.GetPolicyConfigReturns = policyConfig;

    policyConfig.matchType = 'workspaces';
    await act(async () => {
      const c = wrap(
        <PolicyConfigDetails
          clusterName={policyConfig.clusterName}
          name={policyConfig.name}
        />,
      );
      render(c);
    });
    matchedItem.map(item => {
      const AppliedToItem = document.querySelector(
        `span[data-testid="matchItem${item}"]`,
      );
      expect(AppliedToItem).toHaveTextContent(item);
    });
  });

  it('renders Policies', async () => {
    const policyConfig = MockPolicyConfigDetailsResponse.policyConfig;
    const policies = policyConfig.policies;
    api.GetPolicyConfigReturns = policyConfig;

    await act(async () => {
      const c = wrap(
        <PolicyConfigDetails
          clusterName={policyConfig.clusterName}
          name={policyConfig.name}
        />,
      );
      render(c);
    });
    const totalPolicies = document.querySelector(
      'span[data-testid="totalPolicies"]',
    );
    expect(totalPolicies).toHaveTextContent(`(${policyConfig.totalPolicies})`);
    const policiesCard = await screen.getAllByTestId('list-item');
    expect(policiesCard).toHaveLength(policyConfig.totalPolicies);

    const warning = async (id: string) =>
      await screen.getByTestId(`warning-icon-${id}`);

    policies.map(policy => {
      if (policy.status === 'OK') {
        expect(
          document.querySelector(`span[data-testid="policyId-${policy.name}"]`),
        ).toHaveTextContent(policy.name);
      } else {
        expect(warning(policy.id)).toBeTruthy();
        expect(
          document.querySelector(`span[data-testid="policyId-${policy.id}"]`),
        ).toHaveTextContent(policy.id);
      }
      Object.entries(policy.parameters || {}).map(param => {
        const lbl = document.querySelector(`span[data-testid="${param[0]}"]`);
        const value = document.querySelector(
          `span[data-testid="${param[0]}Value"]`,
        );

        expect(lbl).toHaveTextContent(param[0]);
        expect(value).toHaveTextContent(renderParameterValue(param[1]));
      });
    });
  });
});
