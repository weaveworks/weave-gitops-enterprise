import { act, fireEvent, render, screen } from '@testing-library/react';
import moment from 'moment';
import Policies from '..';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import {
  defaultContexts,
  PolicyClientMock,
  TestFilterableTable,
  withContext,
} from '../../../utils/test-utils';

const listPoliciesResponse = {
  policies: [
    {
      category: 'weave.categories.organizational-standards',
      name: 'Prohibit Naked Pods From Being Scheduled',
      id: 'weave.policies.prohibit-naked-pods-from-being-scheduled',
      severity: 'low',
      createdAt: '2022-08-30T11:23:57Z',
      clusterName: 'default/tw-cluster-2',
      tenant: '',
    },
    {
      category: 'weave.categories.organizational-standards',
      name: 'Containers Should Not Run In Namespace',
      id: 'weave.policies.containers-should-not-run-in-namespace',
      severity: 'medium',
      createdAt: '2022-07-30T11:23:55Z',
      clusterName: 'test-dev',
      tenant: 'dev-team',
    },
    {
      category: 'weave.categories.access-control',
      name: 'Rbac Protect Cluster Admin Clusterrolebindings',
      id: 'weave.policies.rbac-protect-cluster-admin-clusterrolebindings',
      severity: 'high',
      createdAt: '2022-08-30T11:23:58Z',
      clusterName: 'default/tw-cluster-2',
      tenant: '',
    },
  ],
  total: 3,
  errors: [],
};
const mappedPolicies = (policies: Array<any>) => {
  return policies.map(e => [
    e.name,
    e.category,
    e.modes.reduce((prev: string, nex: string) => {
      const mode = nex.toLocaleLowerCase() === 'admission' ? 'enforce' : nex;
      return prev + mode;
    }, '') || '-',
    e.tenant || '-',
    e.severity,
    e.clusterName,
    moment(e.createdAt).fromNow(),
  ]);
};
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
        {
          clusterName: 'default/tw-test-cluster',
          namespace: '',
          message: 'second Error message',
        },
      ],
    };

    await act(async () => {
      const c = wrap(<Policies />);
      render(c);
    });

    // TODO "Move Error tests to shared Test"

    const alertMessage = screen.queryByTestId('error-message');
    expect(alertMessage).toHaveTextContent(
      'no matches for kind "Policy" in version "pac.weave.works/v2beta1"',
    );

    // Next Error
    const nextError = screen.queryByTestId('nextError');
    nextError?.click();

    expect(alertMessage).toHaveTextContent('second Error message');

    // Prev error
    const prevError = screen.queryByTestId('prevError');
    prevError?.click();

    expect(alertMessage).toHaveTextContent(
      'no matches for kind "Policy" in version "pac.weave.works/v2beta1"',
    );

    // Error Count
    const errorCount = screen.queryByTestId('errorsCount');
    expect(errorCount?.textContent).toEqual('2');
  });
  it('renders a list of policies', async () => {
    const filterTable = new TestFilterableTable('policy-list', fireEvent);
    api.ListPoliciesReturns = listPoliciesResponse;

    await act(async () => {
      const c = wrap(<Policies />);
      render(c);
    });

    expect(await screen.findByText('Policies')).toBeTruthy();

    filterTable.testRenderTable(
      [
        'Policy Name',
        'Category',
        'Mode',
        'Tenant',
        'Severity',
        'Cluster',
        'Age',
      ],
      listPoliciesResponse.policies.length,
    );

    const search = listPoliciesResponse.policies[0].name;
    const searchedRows = mappedPolicies(
      listPoliciesResponse.policies.filter(e => e.name === search),
    );

    filterTable.testSearchTableByValue(search, searchedRows);
    filterTable.clearSearchByVal(search);
  });
  it('sort policies', async () => {
    api.ListPoliciesReturns = listPoliciesResponse;
    const filterTable = new TestFilterableTable('policy-list', fireEvent);

    await act(async () => {
      const c = wrap(<Policies />);
      render(c);
    });

    expect(await screen.findByText('Policies')).toBeTruthy();

    const sortRowsBySeverity = mappedPolicies(
      listPoliciesResponse.policies.sort((a, b) =>
        a.severity.localeCompare(b.severity),
      ),
    );

    filterTable.testSorthTableByColumn('Severity', sortRowsBySeverity);

    const sortRowsByAge = mappedPolicies(
      listPoliciesResponse.policies.sort(({ createdAt }) => {
        const t = new Date(createdAt).getTime();
        return t * 1;
      }),
    );

    filterTable.testSorthTableByColumn('Age', sortRowsByAge);
  });
});