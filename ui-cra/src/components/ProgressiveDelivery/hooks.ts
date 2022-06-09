import { RequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import { useQuery } from 'react-query';
import { ListCanariesResponse } from '../../cluster-services/prog.pb';
// import { useProgressiveDelivery } from '../../contexts/ProgressiveDelivery';

const CanaryQueryKey = 'canaries';

export function useListCanaries() {
  // const api = useProgressiveDelivery();

  return useQuery<ListCanariesResponse, RequestError>(
    [CanaryQueryKey],
    () => fetch('/v1/canaries').then(res => res.json()),
    // api.ListCanaries({ clusterName: 'somecluster' }),
  );
}
