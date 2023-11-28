import { act, fireEvent, render, screen } from '@testing-library/react';
import moment from 'moment';
import { WorkspaceRoleBindingSubject } from '../../../cluster-services/cluster_services.pb';
import {
  defaultContexts,
  TestFilterableTable,
  withContext,
  WorkspaceClientMock,
} from '../../../utils/test-utils';
import WorkspaceDetails from '../WorkspaceDetails';
import { PoliciesTab } from '../WorkspaceDetails/Tabs/Policies';
import { RoleBindingsTab } from '../WorkspaceDetails/Tabs/RoleBindings';
import { RolesTab } from '../WorkspaceDetails/Tabs/Roles';
import { ServiceAccountsTab } from '../WorkspaceDetails/Tabs/ServiceAccounts';
import { EnterpriseClientContext } from '../../../contexts/API';

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

const MockRoles = {
  name: 'service One',
  clusterName: 'Managemnet',
  objects: [
    {
      name: 'Role one',
      namespacs: 'flux',
      rules: [{ groups: ['group1'], resources: ['res'], verbs: ['verb'] }],
      manifest: '',
      timestamp: '2022-07-30T11:23:55Z',
    },
  ],
  total: 1,
  errors: [],
};
const mappedRoles = (roles: Array<any>) => {
  return roles.map(e => [
    e.name,
    e.namespace || '-',
    'View Rules',
    moment(e.timestamp).fromNow(),
  ]);
};

const MockRoleBinding = {
  name: 'Role Binding one',
  clusterName: 'Managemnet',
  objects: [
    {
      name: 'Role Binding one',
      namespacs: 'flux',
      manifest: '',
      timestamp: '2022-07-30T11:23:55Z',
      role: { apiGroups: 'group1', kind: 'res', name: 'verb' },
      subjects: [
        {
          apiGroup: 'test apiGroup',
          kind: 'test kind',
          name: 'test name',
          namespace: 'flux',
        },
      ],
    },
  ],
  total: 1,
  errors: [],
};
const mappedRoleBinding = (roleBinding: Array<any>) => {
  return roleBinding.map(e => [
    e.name,
    e.namespace || '-',
    e.subjects.map((item: WorkspaceRoleBindingSubject) => item.name).join(', '),
    e.role.name,
    moment(e.timestamp).fromNow(),
  ]);
};

const MockPolicies = {
  name: 'Role Binding one',
  clusterName: 'Managemnet',
  objects: [
    {
      id: '1',
      name: 'Role Binding one',
      category: 'cat One',
      severity: 'High',
      timestamp: '2022-07-30T11:23:55Z',
    },
  ],
  total: 1,
  errors: [],
};
const mappedPolicies = (policy: Array<any>) => {
  return policy.map(e => [
    e.name,
    e.category || '-',
    e.severity,
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
      [EnterpriseClientContext.Provider, { value: { enterprise: api } }],
    ]);
  });

  it('renders service accounts tab', async () => {
    const filterTable = new TestFilterableTable(
      'service-accounts-list',
      fireEvent,
    );
    const workspace = MockWorkspaceResponse.workspace;
    const ListServiceAccounts = MockServiceAccounts.objects;
    api.GetWorkspaceServiceAccountsReturns = MockServiceAccounts;
    await act(async () => {
      const c = wrap(
        <ServiceAccountsTab
          clusterName={workspace.clusterName}
          workspaceName={workspace.name}
        />,
      );
      render(c);
    });

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

  it('renders Roles tab', async () => {
    const filterTable = new TestFilterableTable('roles-list', fireEvent);
    const workspace = MockWorkspaceResponse.workspace;
    const ListRoles = MockRoles.objects;
    api.GetWorkspaceRolesReturns = MockRoles;
    await act(async () => {
      const c = wrap(
        <RolesTab
          clusterName={workspace.clusterName}
          workspaceName={workspace.name}
        />,
      );
      render(c);
    });

    filterTable.testRenderTable(
      ['Name', 'Namespace', 'Rules', 'Age'],
      ListRoles.length,
    );

    const sortRowsByAge = mappedRoles(
      ListRoles.sort((a, b) => {
        const t1 = new Date(a.timestamp).getTime();
        const t2 = new Date(b.timestamp).getTime();
        return t2 - t1;
      }),
    );

    filterTable.testSorthTableByColumn('Age', sortRowsByAge);
  });

  it('renders Role Binding tab', async () => {
    const filterTable = new TestFilterableTable(
      'role-bindings-list',
      fireEvent,
    );
    const workspace = MockWorkspaceResponse.workspace;
    const ListRoleBinding = MockRoleBinding.objects;
    api.GetWorkspaceRoleBindingsReturns = MockRoleBinding;
    await act(async () => {
      const c = wrap(
        <RoleBindingsTab
          clusterName={workspace.clusterName}
          workspaceName={workspace.name}
        />,
      );
      render(c);
    });

    filterTable.testRenderTable(
      ['Name', 'Namespace', 'Bindings', 'Role', 'Age'],
      ListRoleBinding.length,
    );

    const sortRowsByAge = mappedRoleBinding(
      ListRoleBinding.sort((a, b) => {
        const t1 = new Date(a.timestamp).getTime();
        const t2 = new Date(b.timestamp).getTime();
        return t2 - t1;
      }),
    );

    filterTable.testSorthTableByColumn('Age', sortRowsByAge);
  });

  it('renders Policies tab', async () => {
    const filterTable = new TestFilterableTable(
      'workspace-policy-list',
      fireEvent,
    );
    const workspace = MockWorkspaceResponse.workspace;
    const ListPolicies = MockPolicies.objects;
    api.GetWorkspacePoliciesReturn = MockPolicies;
    await act(async () => {
      const c = wrap(
        <PoliciesTab
          clusterName={workspace.clusterName}
          workspaceName={workspace.name}
        />,
      );
      render(c);
    });

    filterTable.testRenderTable(
      ['Name', 'Category', 'Severity', 'Age'],
      ListPolicies.length,
    );

    const sortRowsByAge = mappedPolicies(
      ListPolicies.sort((a, b) => {
        const t1 = new Date(a.timestamp).getTime();
        const t2 = new Date(b.timestamp).getTime();
        return t2 - t1;
      }),
    );

    filterTable.testSorthTableByColumn('Age', sortRowsByAge);
  });
});
