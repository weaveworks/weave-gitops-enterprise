import { ReactQueryOptions } from '@weaveworks/weave-gitops/ui/lib/types';
import { useContext } from 'react';
import { useQuery } from 'react-query';
import {
  ListSopsKustomizationsRequest,
  ListSopsKustomizationsResponse
} from '../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';
import { RequestError } from '../types/custom';


export function useListKustomizationSOPS(
  req: ListSopsKustomizationsRequest,
  opts: ReactQueryOptions<ListSopsKustomizationsResponse, RequestError> = {
    retry: true,
    refetchInterval: 30000,
  },
) {
  const { api } = useContext(EnterpriseClientContext);
  return useQuery<ListSopsKustomizationsResponse, RequestError>(
    ['list_sops', req.clusterName || ''],
    () => api.ListSopsKustomizations(req),
    opts,
  );
}
