import { act, render, screen } from '@testing-library/react';
import Policies from '..';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import {
  defaultContexts,
  PolicyClientMock,
  sortTableByColumn,
  withContext,
} from '../../../utils/test-utils';

describe('ListPolicies', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: PolicyClientMock;

  beforeEach(() => {
    api = new PolicyClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [EnterpriseClientProvider, { api }],
    ]);
  });
  it('renders list policies errors', async () => {
    api.ListPoliciesReturns = {
      policies: [],
      total: 0,
      errors: [
        {
          clusterName: 'default/tw-test-cluster',
          namespace: '',
          message:
            'no matches for kind "Policy" in version "pac.weave.works/v2beta1"',
        },
      ],
    };

    await act(async () => {
      const c = wrap(<Policies />);
      render(c);
    });

    expect(
      await screen.findByText(
        'There were errors while listing some resources:',
      ),
    ).toBeTruthy();

    const alert = document.querySelector('#alert-list-errors');
    const alerts = alert?.querySelectorAll('#alert-list-errors li');

    expect(alerts).toHaveLength(1);
  });
  it('renders a list of policies', async () => {
    api.ListPoliciesReturns = {
      policies: [
        {
          name: 'Containers Running With Privilege Escalation',
          id: 'weave.policies.containers-running-with-privilege-escalation',
          category: 'weave.categories.pod-security',
          severity: 'high',
          createdAt: '2022-06-22T15:54:11Z',
          clusterName: 'management',
          tenant: '',
        },
        {
          name: 'dev-team allowed clusters',
          id: 'weave.policies.tenancy.dev-team-allowed-clusters',
          severity: 'low',
          category: 'weave.categories.tenancy',
          createdAt: '2022-08-19T12:27:14Z',
          clusterName: 'management',
          tenant: 'dev-team',
        },
      ],
      total: 2,
      errors: [
        {
          clusterName: 'default/tw-test-cluster',
          namespace: '',
          message:
            'no matches for kind "Policy" in version "pac.weave.works/v2beta1"',
        },
      ],
    };

    await act(async () => {
      const c = wrap(<Policies />);
      render(c);
    });

    expect(await screen.findByText('Policies')).toBeTruthy();

    const tbl = document.querySelector('#policy-list table');
    const rows = tbl?.querySelectorAll('tbody tr');

    expect(rows).toHaveLength(2);
  });

  it('sort policies by age', async () => {
    api.ListPoliciesReturns = {
      policies: [
        {
          name: 'Containers Running With Privilege Escalation',
          id: 'weave.policies.containers-running-with-privilege-escalation',
          category: 'weave.categories.pod-security',
          severity: 'high',
          createdAt: '2022-06-22T15:54:11Z',
          clusterName: 'management',
          tenant: '',
        },
        {
          name: 'dev-team allowed clusters',
          id: 'weave.policies.tenancy.dev-team-allowed-clusters',
          severity: 'low',
          category: 'weave.categories.tenancy',
          createdAt: '2022-06-19T12:27:14Z',
          clusterName: 'management',
          tenant: 'dev-team',
        },
      ],
      total: 2,
      errors: [],
    };

    await act(async () => {
      const c = wrap(<Policies />);
      render(c);
    });

    expect(await screen.findByText('Policies')).toBeTruthy();

    sortTableByColumn('policy-list', 'Age');

    const text = document.querySelector(
      '#policy-list table tbody tr td',
    )?.textContent;
    expect(text).toMatch('Containers Running With Privilege Escalation');
  });
  it('sort policies by severity', async () => {
    api.ListPoliciesReturns = {
      policies: [
        {
          name: 'Containers Running With Privilege Escalation',
          id: 'weave.policies.containers-running-with-privilege-escalation',
          category: 'weave.categories.pod-security',
          severity: 'low',
          createdAt: '2022-06-22T15:54:11Z',
          clusterName: 'management',
          tenant: '',
        },
        {
          name: 'dev-team allowed clusters',
          id: 'weave.policies.tenancy.dev-team-allowed-clusters',
          severity: 'high',
          category: 'weave.categories.tenancy',
          createdAt: '2022-06-19T12:27:14Z',
          clusterName: 'management',
          tenant: 'dev-team',
        },
      ],
      total: 2,
      errors: [],
    };

    await act(async () => {
      const c = wrap(<Policies />);
      render(c);
    });

    expect(await screen.findByText('Policies')).toBeTruthy();

    sortTableByColumn('policy-list', 'Severity');

    const text = document.querySelector(
      '#policy-list table tbody tr td',
    )?.textContent;
    expect(text).toMatch('dev-team allowed clusters');
  });
});
