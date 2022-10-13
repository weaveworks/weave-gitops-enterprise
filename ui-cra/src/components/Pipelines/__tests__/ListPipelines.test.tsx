import { act, fireEvent, render, screen } from '@testing-library/react';
import Pipelines from '..';
import { PipelinesProvider } from '../../../contexts/Pipelines';
import {
  defaultContexts,
  PipelinesClientMock,
  TestFilterableTable,
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
    const filterTable = new TestFilterableTable('pipelines-list', fireEvent);

    api.ListPipelinesReturns = pipelines;
    const pls = pipelines.pipelines;

    await act(async () => {
      const c = wrap(<Pipelines />);
      render(c);
    });

    expect(await screen.findByText('Pipelines')).toBeTruthy();

    // Check rendered Column header
    filterTable.testRenderTable(
      [
        'Pipeline Name',
        'Pipeline Namespace',
        'Application',
        'Type',
        'Environments',
      ],
      pls.length,
    );

    const search = 'test pipline 2';
    const searchedRows = pls
      .filter(e => e.name === search)
      .map(e => [
        e.name,
        e.namespace,
        e.appRef.name,
        e.appRef.kind,
        e.environments.reduce((prev, nex) => {
          return prev + nex.name;
        }, ''),
      ]);

    filterTable.testSearchTableByValue(search, searchedRows);
    filterTable.clearSearchByVal(search);
  });
});
