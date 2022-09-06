import { act, fireEvent, render, screen } from '@testing-library/react';
import Pipelines from '..';
import { PipelinesProvider } from '../../../contexts/Pipelines';
import {
  PipelinesClientMock,
  withContext,
  defaultContexts,
  TestFilterableTable,
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
  ],
};

const fitlerTabale = new TestFilterableTable('pipelines-list', fireEvent);
const testRowValues = (
  rowValue: NodeListOf<Element>,
  matches: Array<string>,
) => {
  for (let index = 0; index < rowValue.length; index++) {
    const element = rowValue[index];
    expect(element.textContent).toEqual(matches[index]);
  }
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

    const { rows, headers } = fitlerTabale.getTableInfo();

    expect(headers).toHaveLength(4);
    expect(rows).toHaveLength(2);

    testRowValues(headers!, [
      'Pipeline Name',
      'Pipeline Namespace',
      'Type',
      'Environments',
    ]);
  });

  it('search table by pipeline name podinfo', async () => {
    api.ListPipelinesReturns = pipelines;

    await act(async () => {
      const c = wrap(<Pipelines />);
      render(c);
    });

    const { rows } = fitlerTabale.searchTableByValue('podinfo');
    expect(rows).toHaveLength(1);
    const tds = rows![0].querySelectorAll('td');
    testRowValues(tds, [
      'test pipline 2',
      'flux-system',
      'HelmRelease',
      'devprod',
    ]);
  });

  it('filter table by flux-system namespace', async () => {
    api.ListPipelinesReturns = pipelines;

    await act(async () => {
      const c = wrap(<Pipelines />);
      render(c);
    });

    const { rows } = fitlerTabale.applyFilterByValue(
      0,
      'namespace:flux-system',
    );

    expect(rows).toHaveLength(1);
    const tds = rows![0].querySelectorAll('td');

    testRowValues(tds, ['test pipline 2', 'flux-system', 'HelmRelease', 'dev']);
  });
});
