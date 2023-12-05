/* eslint-disable testing-library/no-node-access */
import { act, render, screen } from '@testing-library/react';
import { CoreClientContextProvider } from '@weaveworks/weave-gitops';
import PipelineDetails from '..';
import { GetPipelineResponse } from '../../../../api/pipelines/pipelines.pb';
import { Pipeline } from '../../../../api/pipelines/types.pb';
import { EnterpriseClientContext } from '../../../../contexts/API';
import {
  CoreClientMock,
  PipelinesClientMock,
  defaultContexts,
  withContext,
} from '../../../../utils/test-utils';

const res: GetPipelineResponse = {
  pipeline: {
    name: 'podinfo-02',
    namespace: 'flux-system',
    appRef: {
      apiVersion: 'helm.toolkit.fluxcd.io/v2beta1',
      kind: 'HelmRelease',
      name: 'podinfo',
    },
    environments: [
      {
        name: 'dev',
        promotion: { manual: true, strategy: { notification: {} } },
        targets: [
          {
            namespace: 'podinfo-02-dev',
            clusterRef: {
              kind: 'GitopsCluster',
              name: 'dev',
              namespace: '',
            },
          },
        ],
      },
      {
        name: 'test',
        promotion: { manual: true, strategy: { secretRef: { name: 'test' } } },
        targets: [
          {
            namespace: 'podinfo-02-qa',
            clusterRef: {
              kind: 'GitopsCluster',
              name: 'dev',
              namespace: '',
            },
          },
          {
            namespace: 'podinfo-02-perf',
            clusterRef: {
              kind: 'GitopsCluster',
              name: 'dev',
              namespace: '',
            },
          },
        ],
      },
      {
        name: 'prod',
        promotion: {
          manual: true,
          strategy: {
            pullRequest: {
              type: 'github',
              url: 'https://gitlab.com/weaveworks/cool-project',
              branch: 'main',
            },
          },
        },
        targets: [
          {
            namespace: 'podinfo-02-prod',
            clusterRef: {
              kind: 'GitopsCluster',
              name: 'prod',
              namespace: '',
            },
          },
        ],
      },
    ],
    targets: [],
    status: {
      environments: {
        dev: {
          waitingStatus: { revision: '2.0.0' },
          targetsStatuses: [
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'dev',
                namespace: 'flux-system',
              },
              namespace: 'podinfo-02-dev',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.0.0',
                  lastAppliedRevision: '6.0.0.0',
                  conditions: [
                    {
                      type: 'Ready',
                      status: 'True',
                      reason: 'ReconciliationSucceeded',
                      message: 'Release reconciliation succeeded',
                      timestamp: '2022-12-07T15:06:00Z',
                    },
                    {
                      type: 'Released',
                      status: 'True',
                      reason: 'UpgradeSucceeded',
                      message: 'Helm upgrade succeeded',
                      timestamp: '2022-12-07T15:06:00Z',
                    },
                  ],
                  suspended: false,
                },
              ],
            },
          ],
        },
        prod: {
          targetsStatuses: [
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'prod',
              },
              namespace: 'podinfo-02-prod',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.0.1',
                  lastAppliedRevision: '6.0.1.1',
                  conditions: [
                    {
                      type: 'Ready',
                      status: 'True',
                      reason: 'ReconciliationSucceeded',
                      message: 'Release reconciliation succeeded',
                      timestamp: '2022-09-20T09:06:30Z',
                    },
                    {
                      type: 'Released',
                      status: 'True',
                      reason: 'UpgradeSucceeded',
                      message: 'Helm upgrade succeeded',
                      timestamp: '2022-09-20T09:06:30Z',
                    },
                  ],
                  suspended: false,
                },
              ],
            },
          ],
        },
        test: {
          targetsStatuses: [
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'dev',
                namespace: 'flux-system',
              },
              namespace: 'podinfo-02-qa',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.0.2',
                  lastAppliedRevision: '6.0.2.2',
                  conditions: [
                    {
                      type: 'Ready',
                      status: 'True',
                      reason: 'ReconciliationSucceeded',
                      message: 'Release reconciliation succeeded',
                      timestamp: '2022-09-20T09:07:01Z',
                    },
                    {
                      type: 'Released',
                      status: 'True',
                      reason: 'InstallSucceeded',
                      message: 'Helm install succeeded',
                      timestamp: '2022-09-20T09:07:01Z',
                    },
                  ],
                  suspended: false,
                },
              ],
            },
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'dev',
                namespace: 'flux-system',
              },
              namespace: 'podinfo-02-perf',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.0.3',
                  lastAppliedRevision: '6.0.3.3',
                  conditions: [
                    {
                      type: 'Ready',
                      status: 'True',
                      reason: 'ReconciliationSucceeded',
                      message: 'Release reconciliation succeeded',
                      timestamp: '2022-10-13T13:37:34Z',
                    },
                    {
                      type: 'Released',
                      status: 'True',
                      reason: 'UpgradeSucceeded',
                      message: 'Helm upgrade succeeded',
                      timestamp: '2022-10-13T13:37:34Z',
                    },
                  ],
                  suspended: false,
                },
              ],
            },
          ],
        },
      },
    },
    yaml: 'apiVersion: pipelines.weave.works/v1alpha1\nkind: Pipeline\nmetadata:\n  creationTimestamp: "2022-10-25T16:49:48Z"\n  generation: 4\n  labels:\n    kustomize.toolkit.fluxcd.io/name: pipelines\n    kustomize.toolkit.fluxcd.io/namespace: flux-system\n  managedFields:\n  - apiVersion: pipelines.weave.works/v1alpha1\n    fieldsType: FieldsV1\n    fieldsV1:\n      f:metadata:\n        f:labels:\n          f:kustomize.toolkit.fluxcd.io/name: {}\n          f:kustomize.toolkit.fluxcd.io/namespace: {}\n      f:spec:\n        f:appRef:\n          f:apiVersion: {}\n          f:kind: {}\n          f:name: {}\n        f:environments: {}\n        f:promotion:\n          f:notification: {}\n          f:secretRef:\n            f:name: {}\n    manager: kustomize-controller\n    operation: Apply\n    time: "2022-12-08T13:29:01Z"\n  - apiVersion: pipelines.weave.works/v1alpha1\n    fieldsType: FieldsV1\n    fieldsV1:\n      f:status:\n        f:conditions: {}\n        f:observedGeneration: {}\n    manager: pipeline-controller\n    operation: Update\n    subresource: status\n    time: "2022-10-31T14:30:45Z"\n  name: podinfo-02\n  namespace: flux-system\n  resourceVersion: "230322663"\n  uid: cb61bc7a-9012-4d72-bedd-288699fbfabb\nspec:\n  appRef:\n    apiVersion: helm.toolkit.fluxcd.io/v2beta1\n    kind: HelmRelease\n    name: podinfo\n  environments:\n  - name: dev\n    targets:\n    - clusterRef:\n        kind: GitopsCluster\n        name: dev\n        namespace: flux-system\n      namespace: podinfo-02-dev\n  - name: test\n    targets:\n    - clusterRef:\n        kind: GitopsCluster\n        name: dev\n        namespace: flux-system\n      namespace: podinfo-02-qa\n    - clusterRef:\n        kind: GitopsCluster\n        name: dev\n        namespace: flux-system\n      namespace: podinfo-02-perf\n  - name: prod\n    targets:\n    - clusterRef:\n        kind: GitopsCluster\n        name: prod\n        namespace: default\n      namespace: podinfo-02-prod\n  promotion:\n    notification: {}\nstatus:\n  conditions:\n  - lastTransitionTime: "2022-10-31T14:30:45Z"\n    message: All clusters checked\n    reason: ReconciliationSucceeded\n    status: "True"\n    type: Ready\n  observedGeneration: 4\n',
    type: 'Pipeline',
  },
};

describe('PipelineDetails', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: PipelinesClientMock;
  let core: CoreClientMock;

  beforeEach(() => {
    api = new PipelinesClientMock();
    core = new CoreClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [EnterpriseClientContext.Provider, { value: { pipelines: api } }],
      [CoreClientContextProvider, { api: core }],
    ]);
  });
  it('renders pipeline details', async () => {
    const params = res.pipeline;
    api.GetPipelineReturns = res;
    core.GetObjectReturns = { object: {} };

    await act(async () => {
      const c = wrap(
        <PipelineDetails
          name={params?.name || ''}
          namespace={params?.namespace || ''}
        />,
      );
      render(c);
    });

    // Breadcrumbs
    const title = screen.getByTestId('link-Pipelines').textContent;
    expect(title).toEqual('Pipelines');
    expect(await screen.findByText('podinfo-02')).toBeTruthy();
    //3 envs
    expect(await screen.findByText('dev')).toBeInTheDocument();
    expect(await screen.findByText('test')).toBeInTheDocument();
    expect(await screen.findByText('prod')).toBeInTheDocument();
    //3 targets with applied and speicified versions
    let i = 0;
    while (i <= 3) {
      expect(
        await screen.findByText(`SPECIFIED VERSION: v6.0.${i}`),
      ).toBeInTheDocument();
      expect(
        await screen.findByText(`LAST APPLIED VERSION: v6.0.${i}.${i}`),
      ).toBeInTheDocument();
      i++;
    }
  });

  describe('renders promotion strategy', () => {
    it('pull request', async () => {
      const params = res.pipeline;
      api.GetPipelineReturns = res;
      core.GetObjectReturns = { object: {} };

      await act(async () => {
        const c = wrap(
          <PipelineDetails
            name={params?.name || ''}
            namespace={params?.namespace || ''}
          />,
        );
        render(c);
      });
      expect(screen.getByText('Pull Request')).toBeInTheDocument();
      expect(
        screen.getByText('https://gitlab.com/weaveworks/cool-project'),
      ).toBeInTheDocument();
      expect(screen.getByText('main')).toBeInTheDocument();
      expect(screen.getByText('Secret Ref')).toBeInTheDocument();
      expect(screen.getByText('Notification')).toBeInTheDocument();
    });
  });
  it('handles visibility of promotion button', async () => {
    const params = res.pipeline;
    const manual: Pipeline = {
      ...res.pipeline,
      promotion: {
        manual: true,
      },
    };
    api.GetPipelineReturns = { ...res, pipeline: manual };
    core.GetObjectReturns = { object: {} };

    await act(async () => {
      const c = wrap(
        <PipelineDetails
          name={params?.name || ''}
          namespace={params?.namespace || ''}
        />,
      );
      render(c);
    });
    const buttons = screen.getAllByRole('button');

    const filteredButtons = buttons.filter(e =>
      e.innerHTML.includes('Approve Promotion'),
    );
    expect(filteredButtons.length).toEqual(2);
    const devButton = filteredButtons.filter(
      e => !e.className.includes('Mui-disabled'),
    );
    expect(devButton.length).toEqual(1);
  });
});
