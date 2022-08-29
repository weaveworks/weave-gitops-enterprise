import { useQuery } from 'react-query';
import {
  ListPipelinesResponse,
  Pipelines,
} from '../../api/pipelines/pipelines.pb';

const LIST_PIPLINES_KEY = 'list-piplines';
export const useListPiplines = () => {
  return useQuery<ListPipelinesResponse, Error>(
    [LIST_PIPLINES_KEY],
    () => Pipelines.ListPipelines({}),
    { retry: false },
  );
};
