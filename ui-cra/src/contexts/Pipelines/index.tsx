import React from 'react';
import { useQuery } from 'react-query';
import {
  GetPipelineRequest,
  GetPipelineResponse,
  ListPipelinesResponse,
  Pipelines,
} from '../../api/pipelines/pipelines.pb';
import { formatError } from '../../utils/formatters';
import useNotifications from './../../contexts/Notifications';

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
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<ListPipelinesResponse, Error>(
    [PIPELINES_KEY],
    () => pipelinsService.ListPipelines({}),
    { retry: false, onError },
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
      refetchInterval: 30000,
      retry: false,
      onError,
      enabled,
    },
  );
};
