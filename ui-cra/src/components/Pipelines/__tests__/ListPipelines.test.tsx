import { act, render, screen } from '@testing-library/react';
import Pipelines from '..';
import { PipelinesProvider } from '../../../contexts/Pipelines';
import {
  defaultContexts,
  PipelinesClientMock,
  withContext,
} from '../../../utils/test-utils';

const pipelines = {
  pipelines: [
    {
      name: 'podinfo',
      namespace: 'default',
      appRef: {
        apiVersion: '',
        kind: 'HelmRelease',
        name: 'podinfo',
      },
      environments: [
        {
          name: 'dev',
          targets: [
            {
              namespace: 'podinfo',
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
              namespace: 'podinfo',
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'prod',
              },
            },
          ],
        },
      ],
      targets: [],
    },
    {
      name: 'test pipline 2',
      namespace: 'flux-system',
      appRef: {
        apiVersion: '',
        kind: 'HelmRelease',
        name: 'podinfo 2',
      },
      environments: [
        {
          name: 'dev',
          targets: [
            {
              namespace: 'podinfo',
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'dev',
              },
            },
          ],
        },
      ],
      targets: [],
    },
  ],
};

describe('ListPipelines', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: PipelinesClientMock;

  beforeEach(() => {
    api = new PipelinesClientMock();
    wrap = withContext([...defaultContexts(), [PipelinesProvider, { api }]]);
  });
  it('renders a list of pipelines', async () => {
    api.ListPipelinesReturns = pipelines;
    await act(async () => {
      const c = wrap(<Pipelines />);
      render(c);
    });

    expect(await screen.findByText('Pipelines')).toBeTruthy();

    const tbl = document.querySelector('#pipelines-list table');
    const rows = tbl?.querySelectorAll('tbody tr');

    expect(rows).toHaveLength(pipelines.pipelines.length);
  });
});
