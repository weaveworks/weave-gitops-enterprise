import { coreClient } from '@weaveworks/weave-gitops';
import {
  GetSessionLogsRequest,
  GetSessionLogsResponse,
} from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { useQuery } from 'react-query';
import useNotifications from '../contexts/Notifications';
import { formatError } from '../utils/formatters';
export const useGetLogs = (req: GetSessionLogsRequest) => {
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  const { isLoading, data, error, refetch } = useQuery<
    GetSessionLogsResponse,
    Error
  >([req, 'logs'], () => coreClient.GetSessionLogs(req), {
    retry: false,
    onError,
    refetchInterval: 5000,
  });
  return { isLoading, data, error, refetch };
};
