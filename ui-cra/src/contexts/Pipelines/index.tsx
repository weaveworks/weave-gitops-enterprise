import { useQuery } from 'react-query';
import {
  ListPipelinesResponse,
  Pipelines,
} from '../../api/pipelines/pipelines.pb';

const LIST_PIPLINES_KEY = 'list-piplines';
export const useListPiplines = () => {
  return useQuery<ListPipelinesResponse, Error>(
    [LIST_PIPLINES_KEY],
    () =>
      new Promise((resolve, err) => {
        resolve({
          pipelines: [
            {
              name: 'podinfo',
              namespace: 'default',
              // appRef: {
              //   APIVersion: '',
              //   kind: 'HelmRelease',
              //   name: 'podinfo',
              // },
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
