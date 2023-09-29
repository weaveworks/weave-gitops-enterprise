import React from 'react';
import { useQuery } from 'react-query';
import {
  GetPipelineRequest,
  GetPipelineResponse,
  ListPipelinesResponse,
  ListPullRequestsResponse,
  Pipelines,
} from '../../api/pipelines/pipelines.pb';
import { Pipeline } from '../../api/pipelines/types.pb';
import { formatError } from '../../utils/formatters';
import useNotifications, {
  NotificationData,
} from './../../contexts/Notifications';

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

type SetNotificationsType = React.Dispatch<
  React.SetStateAction<NotificationData[] | []>
>;

const errorHandler =
  (setNotifications: SetNotificationsType) => (error: Error) =>
    setNotifications(formatError(error));

const PIPELINES_KEY = 'pipelines';
export const useListPipelines = () => {
  const pipelinsService = usePipelines();
  const { setNotifications } = useNotifications();

  return useQuery<ListPipelinesResponse, Error>(
    [PIPELINES_KEY],
    () => pipelinsService.ListPipelines({}),
    { retry: false, onError: errorHandler(setNotifications) },
  );
};

export const useGetPipeline = (req: GetPipelineRequest, enabled?: boolean) => {
  const pipelinsService = usePipelines();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<GetPipelineResponse, Error>(
    [PIPELINES_KEY, req.namespace, req.name],
    () => pipelinsService.GetPipeline(req),
    {
      refetchInterval: 5000,
      retry: false,
      onError,
      enabled,
    },
  );
};

export const PIPELINE_PULL_REQUESTS_KEY = 'pipeline_prs';

export const useGetPullRequestsForPipeline = (pipeline?: Pipeline) => {
  const svc = usePipelines();
  const { setNotifications } = useNotifications();

  return useQuery<ListPullRequestsResponse, Error>(
    [PIPELINE_PULL_REQUESTS_KEY, pipeline?.namespace, pipeline?.name],
    () => {
      if (!pipeline) {
        return Promise.resolve(null as any);
      }

      return svc.ListPullRequests({
        pipelineName: pipeline?.name,
        pipelineNamespace: pipeline?.namespace,
      });
    },

    {
      refetchInterval: 30000,
      retry: false,
      onError: errorHandler(setNotifications),
    },
  );
};
