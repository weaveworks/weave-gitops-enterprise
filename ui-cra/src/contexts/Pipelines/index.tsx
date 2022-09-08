import React from 'react';
import { useQuery } from 'react-query';
import {
  GetPipelineRequest,
  GetPipelineResponse,
  ListPipelinesResponse,
  Pipelines,
} from '../../api/pipelines/pipelines.pb';

interface Props {
  api: typeof Pipelines;
  children: any;
}

export const PipelinesContext = React.createContext<typeof Pipelines>(
  null as any,
);

export const PipelinesProvider = ({ api, children }: Props) => (
  <PipelinesContext.Provider value={api}>{children}</PipelinesContext.Provider>
);

export const usePipelines = () => React.useContext(PipelinesContext);

const pList = new Promise<ListPipelinesResponse>((resolve, reject) => {
  resolve({
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
  });
});

const getPipeline = new Promise<GetPipelineResponse>((resolve, reject) => {
  resolve({
    pipeline: {
      name: 'podinfo',
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
              namespace: 'podinfo',
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'management',
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
      status: {
        environments: {
          dev: [
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'management',
              },
              namespace: 'default',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.1.6',
                },
              ],
            },
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'long-cluster-name-that break UI',
              },
              namespace: 'flux-system',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.1.7',
                },
              ],
            },
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'management',
              },
              namespace: 'default',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.1.7',
                },
              ],
            },
          ],
          prod: [
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'management',
              },
              namespace: 'default',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.1.7',
                },
              ],
            },
          ],
          dev2: [
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'management',
              },
              namespace: 'default',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.1.6',
                },
              ],
            },
          ],
          prod2: [
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'management',
              },
              namespace: 'default',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.1.6',
                },
              ],
            },
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'management',
              },
              namespace: 'default',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.1.7',
                },
              ],
            },
          ],
          dev3: [
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'management',
              },
              namespace: 'default',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.1.6',
                },
              ],
            },
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'management',
              },
              namespace: 'default',
              workloads: [
                {
                  kind: 'HelmRelease',
                  name: 'podinfo',
                  version: '6.1.7',
                },
              ],
            },
          ],
          prod3: [
            {
              clusterRef: {
                kind: 'GitopsCluster',
                name: 'management',
              },
              namespace: 'default',
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
      },
    },
  });
});

const LIST_PIPLINES_KEY = 'list-piplines';
export const useListPipelines = () => {
  const pipelinsService = usePipelines();
  return useQuery<ListPipelinesResponse, Error>(
    [LIST_PIPLINES_KEY],
    () => pList, //pipelinsService.ListPipelines({}),
    { retry: false },
  );
};

export const useCountPipelines = () => {
  const { data } = useListPipelines();
  return data?.pipelines?.length;
};

const GET_PIPLINES_KEY = 'get-pipline';
export const useGetPipeline = (req: GetPipelineRequest) => {
  const pipelinsService = usePipelines();
  return useQuery<GetPipelineResponse, Error>(
    [GET_PIPLINES_KEY],
    () => getPipeline,
    { retry: false },
  );
};
