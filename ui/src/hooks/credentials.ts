import { ListCredentialsResponse } from '../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';
import { useContext } from 'react';
import { useQuery } from 'react-query';

export function useListCredentials() {
  const { api } = useContext(EnterpriseClientContext);
  return useQuery<ListCredentialsResponse, Error>('credentials', () =>
    api.ListCredentials({}),
  );
}
