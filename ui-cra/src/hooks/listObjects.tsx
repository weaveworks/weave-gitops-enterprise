import { ListError } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { CoreClientContext } from '@weaveworks/weave-gitops';
import { ReactQueryOptions } from '@weaveworks/weave-gitops/ui/lib/types';
import React from 'react';
import { useQuery } from 'react-query';
import { RequestError } from '../types/custom';

export const useCoreClientContext = () => React.useContext(CoreClientContext);

export function useListObjects<T>(
  type: { new (obj: Object): T },
  kind: string,
  namespace: string = '',
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
    ['list_object', namespace, kind],
    () =>
      api.ListObjects({ namespace, kind }).then(res => {
        const providers = res.objects?.map(obj => new type(obj));
        return { objects: providers, errors: res.errors };
      }),
    opts,
  );
}
