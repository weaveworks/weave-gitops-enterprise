import { ReactQueryOptions } from '@weaveworks/weave-gitops/ui/lib/types';
import { useContext } from 'react';
import { useQuery } from 'react-query';
import {
  ListSopsKustomizationsRequest,
  ListSopsKustomizationsResponse,
} from '../cluster-services/cluster_services.pb';
import { RequestError } from '../types/custom';
import { useAPI } from '../contexts/API';

export function useListKustomizationSOPS(
  req: ListSopsKustomizationsRequest,
  opts: ReactQueryOptions<ListSopsKustomizationsResponse, RequestError> = {
    retry: true,
    refetchInterval: 30000,
  },
) {
  const { enterprise } = useAPI();
  return useQuery<ListSopsKustomizationsResponse, RequestError>(
    ['list_sops', req.clusterName || ''],
    () => enterprise.ListSopsKustomizations(req),
    opts,
  );
}
