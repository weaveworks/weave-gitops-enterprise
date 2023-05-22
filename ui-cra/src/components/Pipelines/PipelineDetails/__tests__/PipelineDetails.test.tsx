/* eslint-disable testing-library/no-node-access */
import { act, render, RenderResult, screen } from '@testing-library/react';
import { CoreClientContextProvider, formatURL } from '@weaveworks/weave-gitops';
import PipelineDetails from '..';
import { GetPipelineResponse } from '../../../../api/pipelines/pipelines.pb';
import { Pipeline } from '../../../../api/pipelines/types.pb';
import { PipelinesProvider } from '../../../../contexts/Pipelines';
import {
  CoreClientMock,
  defaultContexts,
  PipelinesClientMock,
  withContext,
} from '../../../../utils/test-utils';
const fs = require('fs');

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
                  version: '6.2.1',
                  lastAppliedRevision: '6.2.1',
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
                  version: '6.1.6',
                  lastAppliedRevision: '6.1.6',
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
                  version: '6.1.8',
                  lastAppliedRevision: '6.1.8',
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
                  version: '6.1.8',
                  lastAppliedRevision: '6.1.8',
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

interface MappedWorkload {
  kind?: string | undefined;
  name?: string | undefined;
  namespace?: string | undefined;
  version?: string | undefined;
  lastAppliedVersion?: string | undefined;
  mappedClusterName?: string | undefined;
  clusterName?: string | undefined;
}

describe('PipelineDetails', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: PipelinesClientMock;
  let core: CoreClientMock;

  beforeEach(() => {
    api = new PipelinesClientMock();
    core = new CoreClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [PipelinesProvider, { api }],
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
    const breadcrumbs = screen.queryAllByRole('heading');
    expect(breadcrumbs[0].textContent).toEqual('Pipelines');
    expect(await screen.findByText('podinfo-02')).toBeTruthy();

    const targetsStatuses = params?.status?.environments || {};

    // Env & targets
    params?.environments?.forEach(env => {
      const targets = document.querySelectorAll(
        `#${env.name} > [role="targeting"]`,
      );
      expect(targets.length).toEqual(env.targets?.length);

      let workloads: MappedWorkload[] = [];

      targetsStatuses[env.name!].targetsStatuses?.forEach(ts => {
        if (ts.workloads) {
          const wrks = ts.workloads.map(wrk => ({
            ...wrk,
            clusterName: ts.clusterRef?.name || 'management',
            mappedClusterName: ts.clusterRef?.name
              ? `${ts.clusterRef?.namespace || 'default'}/${ts.clusterRef.name}`
              : 'management',
            namespace: ts.namespace,
          }));
          workloads = [...workloads, ...wrks];
        }
      });

      // Targets
      targets.forEach((target, index) => {
        const workloadTarget = target.querySelector('.workloadTarget');

        // Cluster Name
        const clusterNameEle = workloadTarget?.querySelector('.cluster-name');
        checkTextContentToEqual(
          clusterNameEle,
          workloads![index].clusterName || '',
        );

        // Workload Namespace
        const workloadNamespace = workloadTarget?.querySelector(
          '.workload-namespace',
        );
        expect(workloadNamespace?.textContent).toEqual(
          workloads![index].namespace,
        );

        //Target as a link
        const linkToAutomation = target.querySelector('.automation > a');

        const href = formatURL('/helm_release/details', {
          name: workloads![index].name,
          namespace: workloads![index].namespace,
          clusterName: workloads![index].mappedClusterName,
        });
        expect(linkToAutomation).toHaveAttribute('href', href);

        // Workload Last Applied Version
        const lastAppliedRevision = target.querySelector(
          'workloadName > .last-applied-version',
        );
        if (workloads![index].lastAppliedVersion) {
          checkTextContentToEqual(
            lastAppliedRevision,
            workloads![index].lastAppliedVersion || '',
          );
        } else {
          elementToBeNull(lastAppliedRevision);
        }

        // Workload Version
        const workloadVersion = target.querySelector('.version')?.textContent;
        expect(workloadVersion).toEqual(`v${workloads![index].version}`);
      });
    });
  });

  it('renders pipeline Yaml', async () => {
    const params = res.pipeline;
    api.GetPipelineReturns = res;

    await act(async () => {
      const c = wrap(
        <PipelineDetails
          name={params?.name || ''}
          namespace={params?.namespace || ''}
        />,
      );
      render(c);
    });

    const yamlTab = screen
      .getAllByRole('tab')
      .filter(tabEle => tabEle.textContent === 'Yaml')[0];

    yamlTab.click();
    const code = document.querySelector('pre')?.textContent?.trimEnd();
    expect(code).toMatchSnapshot();
  });
  describe('renders promotion strategy', () => {
    it('pull request', async () => {
      const params = res.pipeline;
      const withPromotion: Pipeline = {
        ...res.pipeline,
        promotion: {
          manual: false,
          strategy: {
            pullRequest: {
              type: 'github',
              url: 'https://gitlab.com/weaveworks/cool-project',
              branch: 'main',
            },
          },
        },
      };
      api.GetPipelineReturns = { ...res, pipeline: withPromotion };
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

      const keyVal = document.querySelector('.KeyValueTable');

      expect(keyVal?.textContent).toContain('Pull Request');
      expect(keyVal?.textContent).toContain(
        withPromotion.promotion?.strategy?.pullRequest?.url,
      );
      expect(keyVal?.textContent).toContain(
        withPromotion.promotion?.strategy?.pullRequest?.branch,
      );
      expect(keyVal?.textContent).not.toContain('Notification');
    });
    it('notification', async () => {
      const params = res.pipeline;
      const withPromotion: Pipeline = {
        ...res.pipeline,
        promotion: {
          manual: false,
          strategy: {
            notification: {},
          },
        },
      };
      api.GetPipelineReturns = { ...res, pipeline: withPromotion };
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

      const keyVal = document.querySelector('.KeyValueTable');

      expect(keyVal?.textContent).toContain('Notification');
      expect(keyVal?.textContent).not.toContain('Pull Request');
    });
  });

  describe('snapshots', () => {
    it('renders', async () => {
      const params: any = res.pipeline;
      api.GetPipelineReturns = res;

      let result: RenderResult;
      await act(async () => {
        const c = wrap(
          <PipelineDetails name={params.name} namespace={params.namespace} />,
        );
        result = await render(c);
      });

      //   @ts-ignore
      expect(result.container).toMatchSnapshot();
    });
  });
});

const elementToBeNull = (element: Element | null | undefined) => {
  expect(element).toBeNull();
};

const checkTextContentToEqual = (
  element: Element | null | undefined,
  clusterName: string,
) => {
  expect(element?.textContent).toEqual(clusterName);
};
