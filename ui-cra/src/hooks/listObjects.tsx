import { ListError } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { CoreClientContext } from '@weaveworks/weave-gitops';
import { ListObjectsRequest } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { ReactQueryOptions } from '@weaveworks/weave-gitops/ui/lib/types';
import React, { useContext } from 'react';
import { useQuery } from 'react-query';
import {
  ListSopsKustomizationsRequest,
  ListSopsKustomizationsResponse,
} from '../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';
import { RequestError } from '../types/custom';

export const useCoreClientContext = () => React.useContext(CoreClientContext);

export function useListObjects<T>(
  type: { new (obj: Object): T },
  req: ListObjectsRequest,
  opts: ReactQueryOptions<
    {
      objects?: T[] | undefined;
      errors?: ListError[] | undefined;
    },
    RequestError
  > = {
    retry: true,
    refetchInterval: 30000,
  },
) {
  const { api } = useCoreClientContext();
  return useQuery<
    {
      objects?: T[] | undefined;
      errors?: ListError[] | undefined;
    },
    RequestError
  >(
    ['list_object', req.namespace || '', req.clusterName || '', req.kind],
    () =>
      api.ListObjects(req).then(res => {
        const providers = res.objects?.map(obj => new type(obj));
        return { objects: providers, errors: res.errors };
      }),
    opts,
  );
}

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
