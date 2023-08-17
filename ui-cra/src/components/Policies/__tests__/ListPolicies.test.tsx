import {
  act,
  fireEvent,
  getByTestId,
  render,
  screen,
} from '@testing-library/react';
import moment from 'moment';
import Policies from '../PoliciesListPage';
import {
  defaultContexts,
  CoreClientMock,
  TestFilterableTable,
  withContext,
} from '../../../utils/test-utils';
import { CoreClientContextProvider } from '@weaveworks/weave-gitops';
import { PoliciesTab } from '../PoliciesListTab';

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
      audit: 'audit',
      enforce: '',
    },
    {
      category: 'weave.categories.organizational-standards',
      name: 'Containers Should Not Run In Namespace',
      id: 'weave.policies.containers-should-not-run-in-namespace',
      severity: 'medium',
      createdAt: '2022-07-30T11:23:55Z',
      clusterName: 'test-dev',
      tenant: 'dev-team',
      audit: '',
      enforce: 'enforce',
    },
  ],
  total: 2,
  errors: [],
};
const mappedPolicies = (policies: any[]) => {
  return policies.map(e => [
    e.name,
    e.category,
    e.audit || '-',
    e.enforce || '-',
    e.tenant || '-',
    e.severity,
    e.clusterName,
    moment(e.createdAt).fromNow(),
  ]);
};
describe('ListPolicies', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: CoreClientMock;

  beforeEach(() => {
    api = new CoreClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [CoreClientContextProvider, { api }],
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
  it('renders Tabs of policies Page', async () => {
    api.ListPoliciesReturns = listPoliciesResponse;
    await act(async () => {
      const c = wrap(<Policies />);
      render(c);
    });

    const tabs = await screen.getAllByRole('tab');
    expect(await screen.getByTestId('text-Policies')).toHaveTextContent(
      'Policies',
    );
    expect(tabs).toHaveLength(3);
    expect(tabs[0]).toHaveTextContent('Policies');
    expect(tabs[1]).toHaveTextContent('Policy Audit');
    expect(tabs[2]).toHaveTextContent('Enforcement Events');
  });

  it('renders a list of policies', async () => {
    const filterTable = new TestFilterableTable('policy-list', fireEvent);
    api.ListPoliciesReturns = listPoliciesResponse;
    await act(async () => {
      const c = wrap(<PoliciesTab />);
      render(c);
    });

    filterTable.testRenderTable(
      [
        'Policy Name',
        'Category',
        'Audit',
        'Enforce',
        'Tenant',
        'Severity',
        'Cluster',
        'Age',
      ],
      listPoliciesResponse.policies.length,
    );
  });
});
