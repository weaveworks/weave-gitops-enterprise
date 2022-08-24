import { act, render, screen } from '@testing-library/react';
import PoliciesViolations from '..';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import {
  defaultContexts,
  PolicyClientMock,
  withContext,
} from '../../../utils/test-utils';

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
  it('renders list policy violations errors', async () => {
    api.ListPolicyValidationsReturns = {
      violations: [],
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
      const c = wrap(<PoliciesViolations />);
      render(c);
    });

    expect(
      await screen.findByText(
        'There were errors while listing some resources:',
      ),
    ).toBeTruthy();

    const alert = document.querySelector('#alert-wrapper');
    const alerts = alert?.querySelectorAll('#alert-wrapper li');

    expect(alerts).toHaveLength(1);
  });
  it('renders a list of policy violations', async () => {
    api.ListPolicyValidationsReturns = {
      violations: [
        {
          id: 'e4a12938-660d-439f-96a4-6c70348eda68',
          message:
            'Container Running As User in deployment helm-controller (1 occurrences)',
          clusterId: '659dc1ec-35b4-4d1d-a1de-9371cefcf81e',
          category: 'weave.categories.pod-security',
          severity: 'high',
          createdAt: '2022-08-24T14:08:34Z',
          entity: 'helm-controller',
          namespace: 'flux-system',
          violatingEntity: '',
          name: 'Container Running As User',
          clusterName: 'default/tw-cluster-2',
        },
        {
          id: 'cd364cda-d787-45aa-be88-81fa43c56e63',
          message:
            'Controller ServiceAccount Tokens Automount in deployment helm-controller (1 occurrences)',
          clusterId: '659dc1ec-35b4-4d1d-a1de-9371cefcf81e',
          category: 'weave.categories.access-control',
          severity: 'high',
          createdAt: '2022-08-24T15:58:40Z',
          entity: 'helm-controller',
          namespace: 'flux-system',
          violatingEntity: '',
          name: 'Controller ServiceAccount Tokens Automount',
          clusterName: 'default/tw-cluster-2',
        },
      ],
      total: 2,
      errors: [],
    };

    await act(async () => {
      const c = wrap(<PoliciesViolations />);
      render(c);
    });

    expect(await screen.findByText('Violation Log')).toBeTruthy();

    const tbl = document.querySelector('#violations-list table');
    const rows = tbl?.querySelectorAll('tbody tr');

    expect(rows).toHaveLength(2);
  });

  it('sort policy violations by violated time', async () => {
    api.ListPolicyValidationsReturns = {
      violations: [
        {
          id: 'cd364cda-d787-45aa-be88-81fa43c56e63',
          message:
            'Controller ServiceAccount Tokens Automount in deployment helm-controller (1 occurrences)',
          clusterId: '659dc1ec-35b4-4d1d-a1de-9371cefcf81e',
          category: 'weave.categories.access-control',
          severity: 'high',
          createdAt: '2022-08-24T15:58:40Z',
          entity: 'helm-controller',
          namespace: 'flux-system',
          violatingEntity: '',
          name: 'Controller ServiceAccount Tokens Automount',
          clusterName: 'default/tw-cluster-2',
        },
        {
          id: 'e4a12938-660d-439f-96a4-6c70348eda68',
          message:
            'Container Running As User in deployment helm-controller (1 occurrences)',
          clusterId: '659dc1ec-35b4-4d1d-a1de-9371cefcf81e',
          category: 'weave.categories.pod-security',
          severity: 'high',
          createdAt: '2022-08-24T14:08:34Z',
          entity: 'helm-controller',
          namespace: 'flux-system',
          violatingEntity: '',
          name: 'Container Running As User',
          clusterName: 'default/tw-cluster-2',
        },
      ],
      total: 2,
      errors: [],
    };

    await act(async () => {
      const c = wrap(<PoliciesViolations />);
      render(c);
    });

    expect(await screen.findByText('Violation Log')).toBeTruthy();

    const btns = document.querySelectorAll<HTMLElement>(
      '#violations-list table thead tr th button',
    );
    // Click on Violation Time button
    btns[5].click();
    const text = document.querySelector(
      '#violations-list table tbody tr td',
    )?.textContent;
    expect(text).toMatch(
      'Container Running As User in deployment helm-controller (1 occurrences)',
    );
  });
  it('sort policy violations by severity', async () => {
    api.ListPolicyValidationsReturns = {
      violations: [
        {
          id: 'cd364cda-d787-45aa-be88-81fa43c56e63',
          message:
            'Controller ServiceAccount Tokens Automount in deployment helm-controller (1 occurrences)',
          clusterId: '659dc1ec-35b4-4d1d-a1de-9371cefcf81e',
          category: 'weave.categories.access-control',
          severity: 'low',
          createdAt: '2022-08-24T15:58:40Z',
          entity: 'helm-controller',
          namespace: 'flux-system',
          violatingEntity: '',
          name: 'Controller ServiceAccount Tokens Automount',
          clusterName: 'default/tw-cluster-2',
        },
        {
          id: 'e4a12938-660d-439f-96a4-6c70348eda68',
          message:
            'Container Running As User in deployment helm-controller (1 occurrences)',
          clusterId: '659dc1ec-35b4-4d1d-a1de-9371cefcf81e',
          category: 'weave.categories.pod-security',
          severity: 'high',
          createdAt: '2022-08-24T14:08:34Z',
          entity: 'helm-controller',
          namespace: 'flux-system',
          violatingEntity: '',
          name: 'Container Running As User',
          clusterName: 'default/tw-cluster-2',
        },
      ],
      total: 2,
      errors: [],
    };

    await act(async () => {
      const c = wrap(<PoliciesViolations />);
      render(c);
    });

    expect(await screen.findByText('Violation Log')).toBeTruthy();

    const btns = document.querySelectorAll<HTMLElement>(
      '#violations-list table thead tr th button',
    );
    // Click on Severity button
    btns[3].click();
    const text = document.querySelector(
      '#violations-list table tbody tr td',
    )?.textContent;
    expect(text).toMatch(
      'Container Running As User in deployment helm-controller (1 occurrences)',
    );
  });
});
