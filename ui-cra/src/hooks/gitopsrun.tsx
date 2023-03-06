import { coreClient } from '@weaveworks/weave-gitops';
import { GetSessionLogsResponse } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { useQuery } from 'react-query';
import useNotifications from '../contexts/Notifications';
import { formatError } from '../utils/formatters';

type splitReq = { sessionNamespace: string; sessionId: string; token: string };
export const useGetLogs = (
  req: splitReq,
  logLevelFilter: string,
  logSourceFilter: string,
) => {
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  const { isLoading, data, error, refetch } = useQuery<
    GetSessionLogsResponse,
    Error
  >(
    [logLevelFilter, logSourceFilter, 'logs'],
    () =>
      coreClient.GetSessionLogs({ ...req, logLevelFilter, logSourceFilter }),
    {
      retry: false,
      onError,
      refetchInterval: 5000,
    },
  );
  return { isLoading, data, error, refetch };
};
