import { act, render, screen } from '@testing-library/react';
import Piplines from '..';
import { PipelinesProvider } from '../../../contexts/Pipelines';
import {
  PiplinesClientMock,
  withContext,
  defaultContexts,
} from '../../../utils/test-utils';

describe('ListPiplines', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: PiplinesClientMock;

  beforeEach(() => {
    api = new PiplinesClientMock();
    wrap = withContext([...defaultContexts(), [PipelinesProvider, { api }]]);
  });
  it('renders a list of piplines', async () => {
    api.ListPiplinesReturns = {
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
          name: 'podinfo 2',
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

    await act(async () => {
      const c = wrap(<Piplines />);
      render(c);
    });

    expect(await screen.findByText('Piplines')).toBeTruthy();

    const tbl = document.querySelector('#piplines-list table');
    const headers = tbl?.querySelectorAll('thead tr th');
    const rows = tbl?.querySelectorAll('tbody tr');

    expect(headers).toHaveLength(4);

    expect(headers![0].textContent).toEqual('Name');
    expect(headers![1].textContent).toEqual('Namespace');
    expect(headers![2].textContent).toEqual('Application Name');
    expect(headers![3].textContent).toEqual('Application Kind');

    expect(rows).toHaveLength(2);
  });
});
