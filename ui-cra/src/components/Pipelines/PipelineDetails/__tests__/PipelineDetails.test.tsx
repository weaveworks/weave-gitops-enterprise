/* eslint-disable testing-library/no-node-access */
import { act, render, RenderResult, screen } from '@testing-library/react';
import { CoreClientContextProvider, formatURL } from '@weaveworks/weave-gitops';
import PipelineDetails from '..';
import { GetPipelineResponse } from '../../../../api/pipelines/pipelines.pb';
import { PipelinesProvider } from '../../../../contexts/Pipelines';
import {
  CoreClientMock,
  defaultContexts,
  PipelinesClientMock,
  withContext,
} from '../../../../utils/test-utils';

const res: GetPipelineResponse = {
  pipeline: {
    name: 'podinfo-02',
    namespace: 'default',
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
              namespace: 'flux-system',
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
            },
          },
          {
            namespace: 'podinfo-02-perf',
            clusterRef: {
              kind: 'GitopsCluster',
              name: 'dev',
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
                  version: '6.2.0',
                  lastAppliedRevision: '6.2.0',
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
              },
              namespace: 'podinfo-02-qa',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.1.8',
                },
              ],
            },
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'dev',
              },
              namespace: 'podinfo-02-perf',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.1.8',
                },
              ],
            },
          ],
        },
      },
    },
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
    expect(breadcrumbs[0].textContent).toEqual('Applications');
    expect(breadcrumbs[1].textContent).toEqual('Pipelines');
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
            clusterName: ts.clusterRef?.name,
            mappedClusterName: ts.clusterRef?.namespace
              ? `${ts.clusterRef?.namespace}/${ts.clusterRef.name}`
              : '',
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
        if (workloads![index].clusterName) {
          checkTextContentToEqual(
            clusterNameEle,
            workloads![index].clusterName || '',
          );
        } else {
          elementToBeNull(clusterNameEle);
        }

        // Workload Namespace
        const workloadNamespace = workloadTarget?.querySelector(
          '.workload-namespace',
        );
        expect(workloadNamespace?.textContent).toEqual(
          workloads![index].namespace,
        );

        //Target as a link
        const linkToAutomation = target.querySelector('a');

        if (workloads![index].mappedClusterName) {
          const href = formatURL('/helm_release/details', {
            name: workloads![index].name,
            namespace: workloads![index].namespace,
            clusterName: workloads![index].mappedClusterName,
          });
          linkToExists(linkToAutomation, href);
        }
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

const linkToExists = (element: Element | null, href: string) => {
  expect(element).toHaveAttribute('href', href);
};
const elementToBeNull = (element: Element | null | undefined) => {
  expect(element).toBeNull();
};

const checkTextContentToEqual = (
  element: Element | null | undefined,
  clusterName: string,
) => {
  expect(element?.textContent).toEqual(clusterName);
};
