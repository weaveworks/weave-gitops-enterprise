import { coreClient } from '@weaveworks/weave-gitops';
import {
  IsCRDAvailableResponse,
  ListObjectsResponse,
} from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';

import { useQuery } from 'react-query';
import { RequestError } from '../../utils/test-utils';

export function useListImageAutomation(kind: string, namespace: string = '') {
  return useQuery<ListObjectsResponse, RequestError>(
    ['image_automation', namespace],
    () =>
      coreClient.ListObjects({ namespace, kind }).then(res => {
        const providers = res.objects?.map(obj => obj);
        return { objects: providers, errors: res.errors };
      }),
    { retry: false, refetchInterval: 5000 },
  );
}

export function useCheckCRDInstalled(name: string) {
  return useQuery<boolean, RequestError>(
    ['image_automation_crd_available', name],
    () =>
      coreClient
        .IsCRDAvailable({ name })
        .then(({ clusters }: IsCRDAvailableResponse) => {
          if (!clusters) return false;
          return Object.values(clusters).some(r => r === true);
        }),
    { retry: false, refetchInterval: 5000 },
  );
}
