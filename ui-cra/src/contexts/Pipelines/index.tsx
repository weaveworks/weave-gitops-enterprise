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

const PIPELINES_KEY = 'pipelines';
export const useListPipelines = () => {
  const pipelinsService = usePipelines();
  return useQuery<ListPipelinesResponse, Error>(
    [PIPELINES_KEY],
    () => pipelinsService.ListPipelines({}),
    { retry: false },
  );
};

export const useGetPipeline = (req: GetPipelineRequest) => {
  const pipelinsService = usePipelines();
  return useQuery<GetPipelineResponse, Error>(
    [PIPELINES_KEY, req.namespace, req.name],
    () => Promise.resolve({
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
                  namespace: 'flux-system',
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
                    namespace: 'flux-system',
                  },
                  namespace: 'podinfo-02-dev',
                  workloads: [
                    {
                      kind: 'HelmRelease',
                      name: 'podinfo',
                      version: '6.2.0',
                      lastAppliedRevision: '6.2.0',
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
      }
    }),
    {
      refetchInterval: 30000,
      retry: false,
    },
  );
};
