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

const LIST_PIPLINES_KEY = 'list-piplines';
export const useListPipelines = () => {
  const pipelinsService = usePipelines();
  return useQuery<ListPipelinesResponse, Error>(
    [LIST_PIPLINES_KEY],
    () => pipelinsService.ListPipelines({}),
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
    () => pipelinsService.GetPipeline(req),
    { retry: false },
  );
};
