import { act, fireEvent, render, screen } from '@testing-library/react';
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
    serviceAccount: {
    name: 'service One',
    clusterName: 'Managemnet',
    objects: [],
}
}
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
    const filterTable = new TestFilterableTable('service-accounts-list', fireEvent);
    const workspace = MockWorkspaceResponse.workspace;
    api.ListWSServiceAccountsReturns = MockServiceAccounts.serviceAccount;
    await act(async () => {
      const c = wrap(
        <WorkspaceDetails clusterName="" workspaceName={workspace.name} />,
      );
      render(c);
    });

    const serviceAccountTab = screen
      .getAllByRole('tab')
      .filter(tabEle => tabEle.textContent === 'Service Accounts')[0];
    console.log(serviceAccountTab);
    serviceAccountTab.click();
    // filterTable.testRenderTable(
    //     ['Name', 'Namespaces', 'Cluster'],
    //     MockServiceAccounts.serviceAccount.length,
    //   );
  });
});
