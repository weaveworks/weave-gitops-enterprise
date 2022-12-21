import { act, fireEvent, render, screen } from '@testing-library/react';
import moment from 'moment';
import SecretsList from '..';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import {
  defaultContexts,
  SecretsClientMock,
  TestFilterableTable,
  withContext,
} from '../../../utils/test-utils';

const ListExternalSecretsResponse = {
  secrets: [
    {
      secretName: 'weave.categories.organizational-standards Not Ready',
      externalSecretName: 'Prohibit Naked Pods From Being Scheduled 1-notready',
      namespace: 'weave.policies.prohibit-naked-pods-from-being-scheduled',
      clusterName: 'default/tw-test-cluster',
      secretStore: 'store',
      status: 'Not Ready',
      timestamp: '2022-07-30T11:23:55Z',
    },
    {
      secretName: 'weave.categories.organizational-standards 1',
      externalSecretName: 'Prohibit Naked Pods From Being Scheduled 1',
      namespace: 'weave.policies.prohibit-naked-pods-from-being-scheduled',
      clusterName: 'default/tw-test-cluster',
      secretStore: 'store',
      status: 'Ready',
      timestamp: '2022-08-30T11:23:55Z',
    },
    {
      secretName: 'weave.categories.organizational-standards Ready 2',
      externalSecretName: 'Prohibit Naked Pods From Being Scheduled 2',
      namespace: '',
      clusterName: 'default/tw-test-cluster',
      secretStore: 'store',
      status: 'Ready',
      timestamp: '2022-11-30T11:23:55Z',
    },
  ],
  total: 3,
  errors: [],
};
const mappedSecrets = (secrets: Array<any>) => {
  return secrets.map(e => [
    e.externalSecretName,
    e.status,
    e.namespace || '-',
    e.clusterName,
    e.secretName,
    e.secretStore,
    moment(e.timestamp).fromNow(),
  ]);
};
describe('ListSecrets', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: SecretsClientMock;

  beforeEach(() => {
    api = new SecretsClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [EnterpriseClientProvider, { api }],
    ]);
  });
  it('renders list secrets errors', async () => {
    api.ListSecretsReturns = {
      secrets: [],
      total: 0,
      errors: [
        {
          clusterName: 'default/tw-test-cluster',
          namespace: '',
          message: 'First Error message',
        },
        {
          clusterName: 'default/tw-test-cluster',
          namespace: '',
          message: 'second Error message',
        },
      ],
    };

    await act(async () => {
      const c = wrap(<SecretsList />);
      render(c);
    });

    // TODO "Move Error tests to shared Test"

    const alertMessage = screen.queryByTestId('error-message');
    expect(alertMessage).toHaveTextContent('First Error message');

    // Next Error
    const nextError = screen.queryByTestId('nextError');
    nextError?.click();

    expect(alertMessage).toHaveTextContent('second Error message');

    // Prev error
    const prevError = screen.queryByTestId('prevError');
    prevError?.click();

    expect(alertMessage).toHaveTextContent('First Error message');

    // Error Count
    const errorCount = screen.queryByTestId('errorsCount');
    expect(errorCount?.textContent).toEqual('2');
  });

  it('renders a list of secrets and sort by Name', async () => {
    api.ListSecretsReturns = ListExternalSecretsResponse;
    const secrests = ListExternalSecretsResponse.secrets;

    const filterTable = new TestFilterableTable('secrets-list', fireEvent);

    await act(async () => {
      const c = wrap(<SecretsList />);
      render(c);
    });

    expect(await screen.findByText('Secrets')).toBeTruthy();
    const sortRowsBySecretName = mappedSecrets(
      secrests.sort((a, b) =>
        a.externalSecretName.localeCompare(b.externalSecretName),
      ),
    );
    filterTable.testSorthTableByColumn('Name', sortRowsBySecretName);
  });
  // it('sort Secrets by Age', async () => {
  //   api.ListSecretsReturns = ListExternalSecretsResponse;
  //   const secrests = ListExternalSecretsResponse.secrets;
  //   const filterTable = new TestFilterableTable('secrets-list', fireEvent);
  //   await act(async () => {
  //     const c = wrap(<SecretsList />);
  //     render(c);
  //   });
  //   expect(await screen.findByText('Secrets')).toBeTruthy();

  //   const sortRowsByAge = mappedSecrets(
  //     secrests.sort(({ timestamp }) => {
  //       const t = new Date(timestamp).getTime();
  //       return t * 1;
  //     }),
  //   );

  //   filterTable.testSorthTableByColumn('Age', sortRowsByAge);
  // });
});
