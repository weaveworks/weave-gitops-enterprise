import { act, fireEvent, render, screen } from '@testing-library/react';
import moment from 'moment';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import {
  SecretsClientMock,
  TestFilterableTable,
  defaultContexts,
  withContext,
} from '../../../utils/test-utils';

import SecretDetails from '../SecretDetails';
import { EventsTable } from '../SecretDetails/Events/EventsTable';
const MockSecretResponse = {
  secretName: 'secret Name',
  externalSecretName: 'external Secret Name',
  clusterName: 'managment',
  namespace: 'flux-system',
  secretStore: 'secret Store name',
  secretStoreType: 'secret Store Type',
  secretPath: 'secret Path',
  property: 'property',
  version: 'version',
  status: 'Ok',
  timestamp: '2022-07-30T11:23:55Z',
};
const MockSecretEvents = {
  events: [
    {
      reason: 'Updated Secret to be updated',
      message: 'Updated Secret',
      timestamp: '2022-07-30T11:23:55Z',
      component: 'string',
      host: 'string',
      name: 'string',
      type: 'string',
    },
  ],
  total: 1,
  errors: [],
};

const mappedEvents = (events: Array<any>) => {
  return events.map(e => [e.reason, e.message, moment(e.timestamp).fromNow()]);
};
describe('SecretDetails', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: SecretsClientMock;

  beforeEach(() => {
    api = new SecretsClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [EnterpriseClientProvider, { api }],
    ]);
  });

  it('renders get Secret details with tabs', async () => {
    const secret = MockSecretResponse;
    api.GetExternalSecretReturns = MockSecretResponse;
    await act(async () => {
      const c = wrap(
        <SecretDetails
          externalSecretName={secret.externalSecretName}
          clusterName={secret.clusterName}
          namespace={secret.namespace}
        />,
      );
      render(c);
    });
    //check Tabs
    expect(await screen.getByTitle(secret.externalSecretName)).toBeTruthy();
    const tabs = await screen.getAllByRole('tab');
    expect(secret.clusterName).toBeDefined();
    expect(tabs).toHaveLength(3);
    expect(tabs[0]).toHaveTextContent('Details');
    expect(tabs[1]).toHaveTextContent('Events');
    expect(tabs[2]).toHaveTextContent('Yaml');

    expect(screen.getByTestId('Status')).toHaveTextContent(secret.status);
    expect(screen.getByTestId('Last Updated')).toHaveTextContent(
      moment(secret.timestamp).fromNow(),
    );
    // const namespaces = document.querySelectorAll(
    //   '#workspace-details-header-namespaces span',
    // );
    // expect(namespaces).toHaveLength(workspace.namespaces.length);
  });

  it('renders secret details tab', async () => {
    const secret = MockSecretResponse;
    api.GetExternalSecretReturns = MockSecretResponse;
    await act(async () => {
      const c = wrap(
        <SecretDetails
          externalSecretName={secret.externalSecretName}
          clusterName={secret.clusterName}
          namespace={secret.namespace}
        />,
      );
      render(c);
    });
    expect(screen.getByTestId('External Secret')).toHaveTextContent(
      secret.externalSecretName,
    );
    expect(screen.getByTestId('K8s Secret')).toHaveTextContent(
      secret.secretName,
    );
    expect(screen.getByTestId('Cluster')).toHaveTextContent(secret.clusterName);
    expect(screen.getByTestId('Secret Store')).toHaveTextContent(
      secret.secretStore,
    );
    expect(screen.getByTestId('Secret Store Type')).toHaveTextContent(
      secret.secretStoreType,
    );
    expect(screen.getByTestId('Secret path')).toHaveTextContent(
      secret.secretPath,
    );
    expect(screen.getByTestId('Property')).toHaveTextContent(secret.property);
    expect(screen.getByTestId('Version')).toHaveTextContent(secret.version);
  });

  it('renders events tab', async () => {
    api.ListEventsReturns = MockSecretEvents;
    const filterTable = new TestFilterableTable('events-list', fireEvent);

    await act(async () => {
      const c = wrap(<EventsTable events={MockSecretEvents.events} />);
      render(c);
    });

    filterTable.testRenderTable(
      ['Reason', 'Message', 'Age'],
      MockSecretEvents.events.length,
    );

    const sortRowsByAge = mappedEvents(
      MockSecretEvents.events.sort((a, b) => {
        const t1 = new Date(a.timestamp).getTime();
        const t2 = new Date(b.timestamp).getTime();
        return t2 - t1;
      }),
    );

    filterTable.testSorthTableByColumn('Age', sortRowsByAge);
  });
});
