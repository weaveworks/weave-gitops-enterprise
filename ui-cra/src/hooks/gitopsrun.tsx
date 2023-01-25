import { coreClient } from '@weaveworks/weave-gitops';
import {
  GetSessionLogsRequest,
  GetSessionLogsResponse,
} from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { useQuery } from 'react-query';
export const useGetLogs = (req: GetSessionLogsRequest) => {
  const { isLoading, data, error } = useQuery<GetSessionLogsResponse, Error>(
    'logs',
    () => coreClient.GetSessionLogs(req),
  );
  return { isLoading, data, error };
};
