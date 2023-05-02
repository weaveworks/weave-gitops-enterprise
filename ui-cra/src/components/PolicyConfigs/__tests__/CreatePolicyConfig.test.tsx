import { act, fireEvent, render, screen } from '@testing-library/react';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import RequestContextProvider from '../../../contexts/Request';
import {
  defaultContexts,
  PolicyConfigsClientMock,
  withContext,
  WorkspaceClientMock,
} from '../../../utils/test-utils';
import CreatePolicyConfig from '../create';

const formDataMock = {
  headBranch: 'branch name',
  title: '',
  description: '',
  commitMessage: 'commit',
  repo: 'https://gitlab.git.dev.weave.works/wge/demo-01',
  clusterAutomations: [
    {
      cluster: {
        name: 'management',
        namespace: 'flux-system',
      },
      policyConfig: {
        metadata: {
          name: 'policyConfigName1',
        },
        spec: {
          match: {
            Apps: [],
          },
          config: [],
        },
      },
      isControlPlane: true,
    },
  ],
};

const MockclustersResponse = {
  gitopsClusters: [
    {
      name: 'demo-02',
      namespace: 'default',
      annotations: {},
      labels: {},
      conditions: [
        {
          type: 'Ready',
          status: 'False',
          reason: 'WaitingForSecretDeletion',
          message: 'waiting for access secret to be deleted',
          timestamp: '2023-02-10 09:21:24 +0000 UTC',
        },
        {
          type: 'ClusterConnectivity',
          status: 'False',
          reason: 'ClusterConnectionFailed',
          message:
            'failed connecting to the cluster: Get "https://35.228.134.29/api?timeout=32s": dial tcp 35.228.134.29:443: i/o timeout',
          timestamp: '2022-10-05 19:33:19 +0000 UTC',
        },
      ],
      capiClusterRef: undefined,
      secretRef: { name: 'demo-02-kubeconfig' },
      capiCluster: undefined,
      controlPlane: false,
      type: 'GitopsCluster',
    },
  ],
  total: 1,
  nextPageToken: '',
  errors: [],
};
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

describe('CreatePolicyConfig', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: PolicyConfigsClientMock;
  let WSapi: WorkspaceClientMock;

  let fetch: jest.Mock;

  beforeEach(() => {
    api = new PolicyConfigsClientMock();
    WSapi = new WorkspaceClientMock();

    fetch = jest.fn();

    wrap = withContext([
      ...defaultContexts(),
      [RequestContextProvider, { fetch }],

      [EnterpriseClientProvider, { api }],
    ]);
  });
  it('renders create policyConfig form fields', async () => {
    await act(async () => {
      const c = wrap(<CreatePolicyConfig />);
      render(c);
    });

    expect(await screen.findByText('Create New PolicyConfig')).toBeTruthy();
    expect(
      document.querySelector("input[name='policyConfigName']"),
    ).toBeInTheDocument();

    expect(
      document.querySelector("input[name='clusterName']"),
    ).toBeInTheDocument();

    expect(
      document.querySelector("input[name='matchType']"),
    ).toBeInTheDocument();

    expect(
      document.querySelector("input[name='policies']"),
    ).toBeInTheDocument();
  });

  it('submitting a form', async () => {
    const formData = formDataMock.clusterAutomations[0];
    api.ListGitopsClustersReturns = MockclustersResponse;

    await act(async () => {
      const c = wrap(<CreatePolicyConfig />);
      render(c);
    });

    const policyConfigName = document.querySelector(
      "input[name='policyConfigName']",
    ) as HTMLElement;
    const clusterName = document.querySelector(
      "input[name='clusterName']",
    ) as HTMLElement;

    const matchType = document.querySelector(
      "input[name='matchType']",
    ) as HTMLElement;
    const policies = document.querySelector(
      "input[name='policies']",
    ) as HTMLElement;

    expect(await screen.findByText('Create New PolicyConfig')).toBeTruthy();
    fireEvent.change(policyConfigName, {
      target: { value: 'policyConfigName' },
    });
    console.log((policyConfigName as HTMLInputElement).value);

    fireEvent.change(clusterName, { target: { value: 'management' } });

    fireEvent.change(matchType, { target: { value: 'workspaces' } });
    fireEvent.change(policies, {
      target: {
        value: [
          {
            'weave.policies.containers-minimum-replica-count': {
              parameters: [{ exclude_namespaces: ['2', '4'] }],
            },
          },
          {
            'weave.policies.tenancy.dev-team-allowed-repositories': {
              parameters: [
                {
                  git_urls: [
                    'https://github.com/wkp-example-org/capd-demo-reloade',
                  ],
                },
              ],
            },
          },
        ],
      },
    });

    const form = document.querySelector('form') as HTMLElement;

    await act(async () => {
      fireEvent.submit(form, { target: form });
    });

    expect(fetch).toHaveBeenCalledWith('/v1/enterprise/automations', {
      method: 'POST',
      body: JSON.stringify(formData),
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // await expect(
    //   createDeploymentObjects(
    //     formDataMock,
    //     getProviderToken(GitProvider.GitLab),
    //   ),
    // ).resolves.toHaveProperty('webUrl');
  });

  //   it('test match Type field', async () => {
  //     const formData = formDataMock.clusterAutomations[0];
  //     WSapi.ListWorkspacesReturns = listWorkspacesResponse;
  //     const handleChange = jest.fn();

  //     await act(async () => {
  //       const c = wrap(<CreatePolicyConfig />);
  //       render(c);
  //     });
  //     const matchType = document.querySelector(
  //       "input[name='matchType']",
  //     ) as HTMLElement;
  //     const policyConfigName = document.querySelector(
  //       "input[name='policyConfigName']",
  //     ) as HTMLElement;
  //     fireEvent.click(matchType, {
  //       target: { value: ['Workspaces'] },
  //     });
  //     console.log((matchType as HTMLSelectElement).options);
  //     // expect(await screen.getByTestId('workspacesList')).toBeInTheDocument();
  //   });
});
