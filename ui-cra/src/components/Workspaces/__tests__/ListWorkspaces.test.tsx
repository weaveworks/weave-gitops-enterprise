import { act, fireEvent, render, screen } from '@testing-library/react';
import WorkspacesList from '..';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import {
  defaultContexts,
  WorkspaceClientMock,
  TestFilterableTable,
  withContext,
} from '../../../utils/test-utils';

const listWorkspacesResponse = {
  workspaces: [
    {
      name: 'bar-tenant',
      clusterName: 'management',
      namespaces: ['foo-ns'],
    },
    {
      name: 'foo-tenant',
      clusterName: 'management',
      namespaces: [],
    },
  ],
  total: 2,
  nextPageToken: 'eyJDb250aW51ZVRva2VucyI6eyJtYW5hZ2VtZW50Ijp7IiI6IiJ9fX0K',
  errors: [
    {
      clusterName: 'default/tw-test-cluster',
      namespace: '',
      message:
        'no matches for kind "Workspace" in version "pac.weave.works/v2beta1"',
    },
    {
      clusterName: 'default/tw-test-cluster',
      namespace: '',
      message: 'second Error message',
    },
  ],
};
const mappedWorkspaces = (workspaces: Array<any>) => {
  return workspaces.map(e => [
    e.name,
    e.namespaces?.join(', ') || '-',
    e.clusterName,
  ]);
};
describe('ListWorkspaces', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: WorkspaceClientMock;

  beforeEach(() => {
    api = new WorkspaceClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [EnterpriseClientProvider, { api }],
    ]);
  });
  it('renders list workspaces errors', async () => {
    api.ListWorkspacesReturns = listWorkspacesResponse;

    await act(async () => {
      const c = wrap(<WorkspacesList />);
      render(c);
    });

    // TODO "Move Error tests to shared Test"

    const alertMessage = screen.queryByTestId('error-message');
    expect(alertMessage).toHaveTextContent(
      'no matches for kind "Workspace" in version "pac.weave.works/v2beta1"',
    );

    // Next Error
    const nextError = screen.queryByTestId('nextError');
    nextError?.click();

    expect(alertMessage).toHaveTextContent('second Error message');

    // Prev error
    const prevError = screen.queryByTestId('prevError');
    prevError?.click();

    expect(alertMessage).toHaveTextContent(
      'no matches for kind "Workspace" in version "pac.weave.works/v2beta1"',
    );

    // Error Count
    const errorCount = screen.queryByTestId('errorsCount');
    expect(errorCount?.textContent).toEqual('2');
  });
  it('renders a list of workspaces', async () => {
    const filterTable = new TestFilterableTable('workspaces-list', fireEvent);
    api.ListWorkspacesReturns = listWorkspacesResponse;

    await act(async () => {
      const c = wrap(<WorkspacesList />);
      render(c);
    });

    expect(await screen.findByText('Workspaces')).toBeTruthy();

    filterTable.testRenderTable(
      ['Name', 'Namespaces', 'Cluster'],
      listWorkspacesResponse.workspaces.length,
    );

    const search = listWorkspacesResponse.workspaces[0].name;
    const searchedRows = mappedWorkspaces(
      listWorkspacesResponse.workspaces.filter(e => e.name === search),
    );

    filterTable.testSearchTableByValue(search, searchedRows);
    filterTable.clearSearchByVal(search);

    const sortRowsByName = mappedWorkspaces(
      listWorkspacesResponse.workspaces.sort((a, b) =>
        b.name.localeCompare(a.name),
      ),
    );
    filterTable.testSorthTableByColumn('Name', sortRowsByName);
  });
});
