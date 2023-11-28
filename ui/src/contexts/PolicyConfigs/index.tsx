import { useQuery } from 'react-query';
import {
  GetPolicyConfigRequest,
  GetPolicyConfigResponse,
  ListGitopsClustersRequest,
  ListGitopsClustersResponse,
  ListPolicyConfigsRequest,
  ListPolicyConfigsResponse,
} from '../../cluster-services/cluster_services.pb';
import { formatError } from '../../utils/formatters';
import useNotifications from './../../contexts/Notifications';
import { useAPI } from '../API';

const LIST_ALL_POLICYCONFIS_QUERY_KEY = 'policyConfigs-list';

export function useListPolicyConfigs(req: ListPolicyConfigsRequest) {
  const { enterprise } = useAPI();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<ListPolicyConfigsResponse, Error>(
    [LIST_ALL_POLICYCONFIS_QUERY_KEY, req],
    () => enterprise.ListPolicyConfigs(req),
    { onError },
  );
}

const LIST_POLICYCONFIG_DETAILS_QUERY_KEY = 'policyConfig-details';

export function useGetPolicyConfigDetails(req: GetPolicyConfigRequest) {
  const { enterprise } = useAPI();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<GetPolicyConfigResponse, Error>(
    [LIST_POLICYCONFIG_DETAILS_QUERY_KEY, req],
    () => enterprise.GetPolicyConfig(req),
    { onError },
  );
}

const LIST_CLUSTERS_QUERY_KEY = 'clusters';

export function useGetClustersList(req: ListGitopsClustersRequest) {
  const { enterprise } = useAPI();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<ListGitopsClustersResponse, Error>(
    [LIST_CLUSTERS_QUERY_KEY, req],
    () => enterprise.ListGitopsClusters(req),
    { onError },
  );
}
