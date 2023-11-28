import { useQuery } from 'react-query';
import {
  GetExternalSecretRequest,
  GetExternalSecretResponse,
  ListExternalSecretsRequest,
  ListExternalSecretsResponse,
  ListExternalSecretStoresRequest,
  ListExternalSecretStoresResponse,
} from '../../cluster-services/cluster_services.pb';
import { formatError } from '../../utils/formatters';
import useNotifications from './../../contexts/Notifications';
import { useAPI } from '../API';

const LIST_ALL_SECRETS_QUERY_KEY = 'secrets-list';

export function useListSecrets(req: ListExternalSecretsRequest) {
  const { enterprise } = useAPI();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<ListExternalSecretsResponse, Error>(
    [LIST_ALL_SECRETS_QUERY_KEY, req],
    () => enterprise.ListExternalSecrets(req),
    { onError },
  );
}

const GET_SECRET_QUERY_KEY = 'secret-details';

export function useGetSecretDetails(req: GetExternalSecretRequest) {
  const { enterprise } = useAPI();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<GetExternalSecretResponse, Error>(
    [GET_SECRET_QUERY_KEY, req],
    () => enterprise.GetExternalSecret(req),
    { onError },
  );
}

const List_SECRET_STORE_QUERY_KEY = 'list-secret-store';

export function useListExternalSecretStores(
  req: ListExternalSecretStoresRequest,
) {
  const { enterprise } = useAPI();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<ListExternalSecretStoresResponse, Error>(
    [List_SECRET_STORE_QUERY_KEY, req],
    () => enterprise.ListExternalSecretStores(req),
    { onError },
  );
}
