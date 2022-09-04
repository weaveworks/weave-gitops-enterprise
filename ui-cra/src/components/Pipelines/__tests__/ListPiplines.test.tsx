import { act, fireEvent, render, screen } from '@testing-library/react';
import Pipelines from '..';
import { PipelinesProvider } from '../../../contexts/Pipelines';
import {
  PipelinesClientMock,
  withContext,
  defaultContexts,
  getTableInfo,
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



// WIP - Make a sharable class to test all Filterable table functionality
export function searchTableByValue(tableId: string, searchVal: string) {
  const searchBtn = document.querySelector<HTMLElement>(
    "div[class*='SearchField'] > button",
  );
  searchBtn?.click();
  const searchInput = document.getElementById(
    'table-search',
  ) as HTMLInputElement;

  fireEvent.change(searchInput, { target: { value: searchVal } });

  const searchForm = document.querySelector(
    "div[class*='SearchField'] > form",
  ) as Element;

  fireEvent.submit(searchForm);
  return getTableInfo(tableId);
}

//FilterDialog
export function filterTableByValue(tableId: string) {
  const filterBtn = document.querySelector<HTMLElement>(
    "button[class*='FilterableTable']",
  );
  filterBtn?.click();
  
  const filters = document.querySelectorAll("form > ul > li")

  
  
  return getTableInfo(tableId);
}

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

    const { rows, headers } = getTableInfo('pipelines-list');

    expect(headers).toHaveLength(4);
    expect(headers![0].textContent).toEqual('Name');
    expect(headers![1].textContent).toEqual('Namespace');
    expect(headers![2].textContent).toEqual('Kind');
    expect(headers![3].textContent).toEqual('Environments');

    expect(rows).toHaveLength(2);
  });

  it('search table by pipeline name', async () => {
    api.ListPipelinesReturns = pipelines;

    await act(async () => {
      const c = wrap(<Pipelines />);
      render(c);
    });

    const { rows } = searchTableByValue('pipelines-list', 'podinfo');
    expect(rows).toHaveLength(1);
    const tds = rows![0].querySelectorAll('td');

    expect(tds![0].textContent).toEqual('podinfo');
    expect(tds![1].textContent).toEqual('default');
    expect(tds![2].textContent).toEqual('HelmRelease');
    expect(tds![3].textContent).toContain('dev');
    expect(tds![3].textContent).toContain('prod');
  });
});
