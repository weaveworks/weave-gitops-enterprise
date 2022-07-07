import {
  GetCanaryResponse,
  IsFlaggerAvailableResponse,
  ListCanariesResponse,
  ProgressiveDeliveryService,
} from '@weaveworks/progressive-delivery';
import { useQuery } from 'react-query';
import {
  ClustersService,
  ListEventsRequest,
  ListEventsResponse,
} from '../cluster-services/cluster_services.pb';

interface FlaggerStatus {
  isFlaggerAvailable: boolean;
}

export function useIsFlaggerAvailable() {
  return useQuery<FlaggerStatus, Error>('flagger', () =>
    ProgressiveDeliveryService.IsFlaggerAvailable({}).then(
      ({ clusters }: IsFlaggerAvailableResponse) => {
        return clusters === undefined || Object.keys(clusters).length === 0
          ? { isFlaggerAvailable: false }
          : {
              isFlaggerAvailable: Object.values(clusters).some(
                (value: boolean) => value === true,
              ),
            };
      },
    ),
  );
}
const CANARIES_POLL_INTERVAL = 60000;
export function useListCanaries(clb: (res: ListCanariesResponse) => void) {
  return useQuery<ListCanariesResponse, Error>(
    'canaries',
    () =>
      ProgressiveDeliveryService.ListCanaries({}).then(res => {
        clb(res);
        return res;
      }),
    {
      refetchInterval: CANARIES_POLL_INTERVAL,
    },
  );
}

export function useListCanariesCount() {
  const { data } = useQuery<ListCanariesResponse, Error>(
    'canaries',
    () => ProgressiveDeliveryService.ListCanaries({}),
    {},
  );
  return data?.canaries?.length || 0;
}

export function useListEvents(listEventsPayload: ListEventsRequest) {
  return useQuery<ListEventsResponse, Error>('events', () =>
    ClustersService.ListEvents(listEventsPayload),
  );
}
export interface CanaryParams {
  name: string;
  namespace: string;
  clusterName: string;
}
export function useCanaryDetails({
  name,
  namespace,
  clusterName,
}: CanaryParams) {
  return useQuery<GetCanaryResponse, Error>(
    'canaryDetails',
    () =>
      ProgressiveDeliveryService.GetCanary({
        name,
        namespace,
        clusterName,
      }),
    {},
  );
}
