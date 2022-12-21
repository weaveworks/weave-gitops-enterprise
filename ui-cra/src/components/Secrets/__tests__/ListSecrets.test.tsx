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
      secretName: 'weave.categories.organizational-standards',
      externalSecretName: 'Prohibit Naked Pods From Being Scheduled',
      namespace: 'weave.policies.prohibit-naked-pods-from-being-scheduled',
      clusterName: 'default/tw-test-cluster',
      secretStore: 'store',
      status: 'Not Ready',
      timestamp: '2022-07-30T11:23:55Z',
    },
    {
      secretName: 'weave.categories.organizational-standards',
      externalSecretName: 'Prohibit Naked Pods From Being Scheduled',
      namespace: 'weave.policies.prohibit-naked-pods-from-being-scheduled',
      clusterName: 'default/tw-test-cluster',
      secretStore: 'store',
      status: 'Ready',
      timestamp: '2022-07-30T11:23:55Z',
    },
    {
      secretName: 'weave.categories.organizational-standards',
      externalSecretName: 'Prohibit Naked Pods From Being Scheduled',
      namespace: '',
      clusterName: 'default/tw-test-cluster',
      secretStore: 'store',
      status: 'Ready',
      timestamp: '2022-07-30T11:23:55Z',
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
  it('renders list secret errors', async () => {
    api.ListSecretsReturns = {
      secrets: [],
      total: 0,
      errors: [
        {
          clusterName: 'default/tw-test-cluster',
          namespace: '',
          message:
            'no matches for kind "Secret" in version "pac.weave.works/v2beta1"',
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
    
    expect(alertMessage).toHaveTextContent(
      'no matches for kind "Secret" in version "pac.weave.works/v2beta1"',
    );

    // Next Error
    const nextError = screen.queryByTestId('nextError');
    nextError?.click();

    expect(alertMessage).toHaveTextContent('second Error message');

    // Prev error
    const prevError = screen.queryByTestId('prevError');
    prevError?.click();

    expect(alertMessage).toHaveTextContent(
      'no matches for kind "Secret" in version "pac.weave.works/v2beta1"',
    );

    // Error Count
    const errorCount = screen.queryByTestId('errorsCount');
    expect(errorCount?.textContent).toEqual('2');
  });
  it('renders a list of secrets', async () => {
    api.ListSecretsReturns = ListExternalSecretsResponse;

    await act(async () => {
      const c = wrap(<SecretsList />);
      render(c);
    });

    expect(await screen.findByText('Secrets')).toBeTruthy();

  });
  it('sort policies', async () => {
    api.ListSecretsReturns = ListExternalSecretsResponse;
    const filterTable = new TestFilterableTable('secrets-list', fireEvent);

    await act(async () => {
      const c = wrap(<SecretsList />);
      render(c);
    });

    expect(await screen.findByText('Secrets')).toBeTruthy();

    const sortRowsByName = mappedSecrets(
      ListExternalSecretsResponse.secrets.sort((a, b) =>
        a.externalSecretName.localeCompare(b.externalSecretName),
      ),
    );

    filterTable.testSorthTableByColumn('Name', sortRowsByName);

  });
});