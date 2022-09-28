import {
  act,
  render,
  RenderResult,
  screen,
} from '@testing-library/react';
import PipelineDetails from '..';
import { GetPipelineResponse } from '../../../../api/pipelines/pipelines.pb';
import { PipelinesProvider } from '../../../../contexts/Pipelines';
import {
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
              },
              namespace: 'podinfo-02-dev',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.2.0',
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

describe('PipelineDetails', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: PipelinesClientMock;

  beforeEach(() => {
    api = new PipelinesClientMock();
    wrap = withContext([...defaultContexts(), [PipelinesProvider, { api }]]);
  });
  it('renders pipeline details', async () => {
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

      let workloads: {
        target: string | undefined;
        kind?: string | undefined;
        name?: string | undefined;
        version?: string | undefined;
      }[] = [];

      targetsStatuses[env.name!].targetsStatuses?.forEach(ts => {
        if (ts.workloads) {
          const wrks = ts.workloads.map(wrk => ({
            ...wrk,
            target: ts.clusterRef?.name
              ? `${ts.clusterRef?.name}/${ts.namespace}`
              : ts.namespace,
          }));
          workloads = [...workloads, ...wrks];
        }
      });

      // Targets
      targets.forEach((target, index) => {
        // Target
        const workloadTarget =
          target.querySelector('.workloadTarget')?.textContent;
        expect(workloadTarget).toEqual(workloads![index].target);

        // Workload Name
        const workloadName = target.querySelector('.workloadName')?.textContent;
        expect(workloadName).toEqual(workloads![index].name);

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
