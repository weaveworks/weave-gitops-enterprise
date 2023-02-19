import { IsCRDAvailableResponse } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { ListError } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { CoreClientContext } from '@weaveworks/weave-gitops';
import React from 'react';
import { useQuery } from 'react-query';
import { RequestError } from '../../utils/test-utils';

export const useCoreClientContext = () => React.useContext(CoreClientContext);

export function useListImageObjects<T>(
  type: { new (obj: Object): T },
  kind: string,
  namespace: string = '',
) {
  const { api } = useCoreClientContext();

  return useQuery<
    {
      objects?: T[] | undefined;
      errors?: ListError[] | undefined;
    },
    RequestError
  >(
    ['list_image_automation_object', namespace, kind],
    () =>
      api.ListObjects({ namespace, kind }).then(res => {
        const providers = res.objects?.map(obj => new type(obj));
        return { objects: providers, errors: res.errors };
      }),
    { retry: false, refetchInterval: 30000 },
  );
}
export function useCheckCRDInstalled(name: string) {
  const { api } = useCoreClientContext();
  return useQuery<boolean, RequestError>(
    ['image_automation_crd_available', name],
    () =>
      api
        .IsCRDAvailable({ name })
        .then(({ clusters }: IsCRDAvailableResponse) => {
          if (!clusters) return false;
          return Object.values(clusters).some(r => r === true);
        }),
    { retry: false, refetchInterval: data => (data ? false : 30000) },
  );
}
