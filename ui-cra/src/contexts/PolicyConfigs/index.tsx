import { useContext } from 'react';
import { useQuery } from 'react-query';
import {
    ListPolicyConfigsRequest,
    ListPolicyConfigsResponse
} from '../../cluster-services/cluster_services.pb';
import { formatError } from '../../utils/formatters';
import { EnterpriseClientContext } from '../EnterpriseClient';
import useNotifications from './../../contexts/Notifications';

const LIST_ALL_PolicyConfigs_QUERY_KEY = 'policyConfigs-list';

export function useListPolicyConfigs(req: ListPolicyConfigsRequest) {
  const { api } = useContext(EnterpriseClientContext);
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  
  return useQuery<ListPolicyConfigsResponse, Error>(
    [LIST_ALL_PolicyConfigs_QUERY_KEY, req],
    () => api.ListPolicyConfigs(req),
    { onError },
  );
}
