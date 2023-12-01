import { useContext } from 'react';
import { useQuery } from 'react-query';
import { ListCredentialsResponse } from '../cluster-services/cluster_services.pb';
import { useAPI } from '../contexts/API';

export function useListCredentials() {
  const { clustersService } = useAPI();
  return useQuery<ListCredentialsResponse, Error>('credentials', () =>
    clustersService.vice.ListCredentials({}),
  );
}
