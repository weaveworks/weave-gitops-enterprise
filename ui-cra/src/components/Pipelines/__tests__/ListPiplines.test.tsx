import { act, render, screen } from '@testing-library/react';
import Pipelines from '..';
import { PipelinesProvider } from '../../../contexts/Pipelines';
import {
  PipelinesClientMock,
  withContext,
  defaultContexts,
  getTableInfo,
  searchTableByValue,
} from '../../../utils/test-utils';

describe('ListPipelines', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: PipelinesClientMock;

  beforeEach(() => {
    api = new PipelinesClientMock();
    wrap = withContext([...defaultContexts(), [PipelinesProvider, { api }]]);
  });
  it('renders a list of pipelines', async () => {
    api.ListPipelinesReturns = {
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
          name: 'test pipeline',
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
      const c = wrap(<Pipelines />);
      render(c);
    });

    expect(await screen.findByText('Pipelines')).toBeTruthy();

    const { rows, headers } = getTableInfo('pipelines-list');

    expect(headers).toHaveLength(4);
    expect(headers![0].textContent).toEqual('Name');
    expect(headers![1].textContent).toEqual('Namespace');
    expect(headers![2].textContent).toEqual('Application Name');
    expect(headers![3].textContent).toEqual('Application Kind');

    expect(rows).toHaveLength(2);
  });

  it('search table by pipeline name', async () => {
    api.ListPipelinesReturns = {
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

    await act(async () => {
      const c = wrap(<Pipelines />);
      render(c);
    });

    const { rows } = searchTableByValue('pipelines-list', 'podinfo');
    expect(rows).toHaveLength(1);
  });
});
