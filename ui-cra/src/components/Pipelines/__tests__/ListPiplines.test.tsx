import { act, render, screen } from '@testing-library/react';
import Pipelines from '..';
import { PipelinesProvider } from '../../../contexts/Pipelines';
import {
  PipelinesClientMock,
  withContext,
  defaultContexts,
} from '../../../utils/test-utils';
import { TestFilterableTable } from './FilterableTable.test';

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

const fitlerTabale = new TestFilterableTable('pipelines-list');

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

    fitlerTabale.testRenderTable(
      ['Pipeline Name', 'Pipeline Namespace', 'Type', 'Environments'],
      2,
    );
  });

  it('search table by pipeline name test pipline 2', async () => {
    api.ListPipelinesReturns = pipelines;

    await act(async () => {
      const c = wrap(<Pipelines />);
      render(c);
    });

    fitlerTabale.testSearchTableByValue('test pipline 2', 0, [
      'test pipline 2',
      'flux-system',
      'HelmRelease',
      'dev',
    ]);
  });

  it('filter table by flux-system namespace', async () => {
    api.ListPipelinesReturns = pipelines;

    await act(async () => {
      const c = wrap(<Pipelines />);
      render(c);
    });

    fitlerTabale.testFilterTableByValue(0, 'namespace:flux-system', [
      'test pipline 2',
      'flux-system',
      'HelmRelease',
      'dev',
    ]);
  });
});
