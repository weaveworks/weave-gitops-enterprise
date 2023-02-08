import { act, fireEvent, render, screen } from '@testing-library/react';
import moment from 'moment';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import {
  defaultContexts,
  WorkspaceClientMock,
  withContext,
  TestFilterableTable,
} from '../../../utils/test-utils';
import WorkspaceDetails from '../WorkspaceDetails';
const MockWorkspaceResponse = {
  workspace: {
    name: 'dev-team',
    namespaces: [],
    clusterName: 'management',
  },
};
const MockServiceAccounts = {
  name: 'service One',
  clusterName: 'Managemnet',
  objects: [
    {
      name: 'test 1',
      namespace: 'namespace 1',
      timestamp: '2022-08-30T11:23:57Z',
      manifest: '',
    },
    {
      name: 'test 2',
      namespace: 'namespace 2',
      timestamp: '2022-07-30T11:23:55Z',
      manifest: '',
    },
  ],
  total: 2,
  errors: [],
};
const mappedServiceAccounts = (serviceAccounts: Array<any>) => {
  return serviceAccounts.map(e => [
    e.name,
    e.namespace || '-',
    moment(e.timestamp).fromNow(),
  ]);
};
describe('WorkspaceDetails', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: WorkspaceClientMock;

  beforeEach(() => {
    api = new WorkspaceClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [EnterpriseClientProvider, { api }],
    ]);
  });

  it('renders get Workspace details', async () => {
    const workspace = MockWorkspaceResponse.workspace;
    api.GetWorkspaceReturns = MockWorkspaceResponse.workspace;
    await act(async () => {
      const c = wrap(
        <WorkspaceDetails clusterName="" workspaceName={workspace.name} />,
      );
      render(c);
    });
    expect(await screen.getByTitle(workspace.name)).toBeTruthy();
    expect(workspace.clusterName).toBeDefined();
    expect(await screen.getAllByRole('tab')).toHaveLength(4);

    // Details
    expect(screen.getByTestId('Workspace Name')).toHaveTextContent(
      workspace.name,
    );
    // Namespaces
    const namespaces = document.querySelectorAll(
      '#workspace-details-header-namespaces span',
    );
    expect(namespaces).toHaveLength(workspace.namespaces.length);
  });

  it('renders service accounts tab', async () => {
    const filterTable = new TestFilterableTable(
      'service-accounts-list',
      fireEvent,
    );
    const workspace = MockWorkspaceResponse.workspace;
    const ListServiceAccounts = MockServiceAccounts.objects;
    api.ListWSServiceAccountsReturns = MockServiceAccounts;
    await act(async () => {
      const c = wrap(
        <WorkspaceDetails clusterName="" workspaceName={workspace.name} />,
      );
      render(c);
    });

    const serviceAccountTab = screen
      .getAllByRole('tab')
      .filter(tabEle => tabEle.textContent === 'Service Accounts')[0];
    serviceAccountTab.click();

    filterTable.testRenderTable(
      ['Name', 'Namespace', 'Age'],
      ListServiceAccounts.length,
    );

    const sortRowsByAge = mappedServiceAccounts(
      ListServiceAccounts.sort((a, b) => {
        const t1 = new Date(a.timestamp).getTime();
        const t2 = new Date(b.timestamp).getTime();
        return t2 - t1;
      }),
    );

    filterTable.testSorthTableByColumn('Age', sortRowsByAge);
  });
});
