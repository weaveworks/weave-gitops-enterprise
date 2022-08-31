import React from 'react';
import { useQuery } from 'react-query';
import {
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

const LIST_PIPLINES_KEY = 'list-pipelines';
export const useListPipelines = () => {
  const pipelinsService = usePipelines();
  return useQuery<ListPipelinesResponse, Error>(
    [LIST_PIPLINES_KEY],
    () =>
      new Promise((resolve, err) => {
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
              name: 'pipeline test',
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
        });
      }),
    { retry: false },
  );
};
