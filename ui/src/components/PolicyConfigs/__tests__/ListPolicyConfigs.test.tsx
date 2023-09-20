import { act, fireEvent, render, screen } from '@testing-library/react';
import PolicyConfigsList from '..';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import {
  defaultContexts,
  PolicyConfigsClientMock,
  TestFilterableTable,
  withContext,
} from '../../../utils/test-utils';
import moment from 'moment';

const listPolicyConfigsResponse = {
  policyConfigs: [
    {
      name: 'PolicyConfig number one',
      clusterName: 'test-dev',
      totalPolicies: 1,
      status: 'Ok',
      match: 'application',
      age: '2022-08-30T11:23:55Z',
    },
    {
      name: 'name two',
      clusterName: 'test-dev',
      totalPolicies: 2,
      status: 'Ok',
      match: 'application',
      age: '2022-07-30T11:23:55Z',
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
    {
      clusterName: 'default/tw-test-cluster',
      namespace: '',
      message: 'second Error message',
    },
  ],
};
const mappedPolicyConfigs = (policyConfigs: Array<any>) => {
  return policyConfigs.map(e => [
    ' ',
    e.name,
    e.clusterName,
    e.totalPolicies.toString(),
    e.match,
    moment(e.age).fromNow(),
  ]);
};
describe('ListPolicyConfigs', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: PolicyConfigsClientMock;

  beforeEach(() => {
    api = new PolicyConfigsClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [EnterpriseClientProvider, { api }],
    ]);
  });
  it('renders list policies errors', async () => {
    api.ListPolicyConfigsReturns = listPolicyConfigsResponse;

    await act(async () => {
      const c = wrap(<PolicyConfigsList />);
      render(c);
    });

    expect(await screen.findByText('PolicyConfigs')).toBeTruthy();

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

  it('renders a list of policyConfigs', async () => {
    const filterTable = new TestFilterableTable(
      'policyConfigs-list',
      fireEvent,
    );
    api.ListPolicyConfigsReturns = listPolicyConfigsResponse;
    const policyConfigs = listPolicyConfigsResponse.policyConfigs;

    await act(async () => {
      const c = wrap(<PolicyConfigsList />);
      render(c);
    });

    expect(await screen.findByText('PolicyConfigs')).toBeTruthy();

    filterTable.testRenderTable(
      ['', 'Name', 'Cluster', 'Policy Count', 'Applied To', 'Age'],
      policyConfigs.length,
    );

    const search = listPolicyConfigsResponse.policyConfigs[0].name;
    const searchedRows = mappedPolicyConfigs(
      policyConfigs.filter(e => e.name === search),
    );

    filterTable.testSearchTableByValue(search, searchedRows);
    filterTable.clearSearchByVal(search);
  });
  it('sort policyConfigs', async () => {
    const filterTable = new TestFilterableTable(
      'policyConfigs-list',
      fireEvent,
    );
    api.ListPolicyConfigsReturns = listPolicyConfigsResponse;
    const policyConfigs = listPolicyConfigsResponse.policyConfigs;
    await act(async () => {
      const c = wrap(<PolicyConfigsList />);
      render(c);
    });

    const sortRowsByName = mappedPolicyConfigs(
      policyConfigs.sort((a, b) => a.name.localeCompare(b.name)),
    );
    filterTable.testSorthTableByColumn('Name', sortRowsByName);
    const sortRowsByAge = mappedPolicyConfigs(
      policyConfigs.sort((a, b) => {
        const t1 = new Date(a.age).getTime();
        const t2 = new Date(b.age).getTime();
        return t2 - t1;
      }),
    );
    filterTable.testSorthTableByColumn('Age', sortRowsByAge);
  });
  it('renders Warning Icon', async () => {
    listPolicyConfigsResponse.policyConfigs[0].status = 'Warning';
    api.ListPolicyConfigsReturns = listPolicyConfigsResponse;
    const name = listPolicyConfigsResponse.policyConfigs[0].name;
    await act(async () => {
      const c = wrap(<PolicyConfigsList />);
      render(c);
    });
    expect(screen.getByTestId(`warning-icon-${name}`)).toBeTruthy();
  });
});
