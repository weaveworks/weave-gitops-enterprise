import { act, render, RenderResult, screen } from '@testing-library/react';
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
    name: 'podinfo-pipeline',
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
            namespace: 'dev',
            clusterRef: {
              kind: '',
              name: '',
            },
          },
        ],
      },
      {
        name: 'prod',
        targets: [
          {
            namespace: 'prod',
            clusterRef: {
              kind: '',
              name: '',
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
                kind: '',
                name: '',
              },
              namespace: 'default',
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
        prod: {
          targetsStatuses: [
            {
              clusterRef: {
                kind: '',
                name: '',
              },
              namespace: 'default',
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
  it('renders a list of pipelines', async () => {
    const params: any = res.pipeline;
    api.GetPipelineReturns = res;

    await act(async () => {
      const c = wrap(
        <PipelineDetails
          name={params.name}
          namespace={params.namespace}
        />,
      );
      render(c);
    });

    expect(await screen.findByText('prod')).toBeTruthy();
    expect(await screen.findByText('v6.2.0')).toBeTruthy();
    expect(await screen.findByText('dev')).toBeTruthy();
    expect(await screen.findByText('v6.1.8')).toBeTruthy();
  });
  describe('snapshots', () => {
    it('renders', async () => {
      const params: any = res.pipeline;
      api.GetPipelineReturns = res;

      let result: RenderResult;
      await act(async () => {
        const c = wrap(
          <PipelineDetails
            name={params.name}
            namespace={params.namespace}
          />,
        );
        result = await render(c);
      });

      //   @ts-ignore
      expect(result.container).toMatchSnapshot();
    });
  });
});
