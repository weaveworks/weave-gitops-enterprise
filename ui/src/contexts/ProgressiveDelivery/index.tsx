import {
  GetCanaryResponse,
  ListCanariesResponse,
  ListCanaryObjectsResponse,
} from '@weaveworks/progressive-delivery';
import _ from 'lodash';
import React, { useContext } from 'react';
import { useQuery } from 'react-query';
import {
  ListEventsRequest,
  ListEventsResponse,
} from '../../cluster-services/cluster_services.pb';
import { formatError } from '../../utils/formatters';
import { EnterpriseClientContext } from '../EnterpriseClient';
import useNotifications from './../../contexts/Notifications';
import { useAPI } from '../API';

export const useProgressiveDelivery = () => {
  const { progressiveDeliveryService } = useAPI();
  return progressiveDeliveryService;
};
const PD_QUERY_KEY = 'flagger';
const FLAGGER_STATUS_KEY = 'status';

export const useIsFlaggerAvailable = () => {
  const pd = useProgressiveDelivery();

  return useQuery<boolean, Error>([PD_QUERY_KEY, FLAGGER_STATUS_KEY], () => {
    return pd.IsFlaggerAvailable({}).then(res => {
      if (!res.clusters) {
        return false;
      }

      return _.includes(_.values(res.clusters), true);
    });
  });
};

const CANARIES_KEY = 'canaries';
export const useListCanaries = () => {
  const pd = useProgressiveDelivery();

  return useQuery<ListCanariesResponse, Error>(
    [PD_QUERY_KEY, CANARIES_KEY],
    () => pd.ListCanaries({}),
    { refetchInterval: 10000 },
  );
};

export type CanaryParams = {
  name: string;
  namespace: string;
  clusterName: string;
};

const GET_CANARY_KEY = 'get_canary_details';
export const useGetCanaryDetails = (params: CanaryParams) => {
  const pd = useProgressiveDelivery();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<GetCanaryResponse, Error>(
    [PD_QUERY_KEY, GET_CANARY_KEY, params],
    () => pd.GetCanary(params),
    {
      refetchInterval: data => (data ? false : 10000),
      retry: false,
      onError,
    },
  );
};

const CANARY_OBJS_KEY = 'canary_objects';
export const useListFlaggerObjects = (params: CanaryParams) => {
  const pd = useProgressiveDelivery();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<ListCanaryObjectsResponse, Error>(
    [PD_QUERY_KEY, CANARY_OBJS_KEY, params],
    () => {
      return pd.ListCanaryObjects(params);
    },
    {
      onError,
    },
  );
};

const EVENTS_QUERY_KEY = 'events';

export function useListEvents(req: ListEventsRequest) {
  const { api } = useContext(EnterpriseClientContext);
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<ListEventsResponse, Error>(
    [EVENTS_QUERY_KEY, req],
    () => api.ListEvents(req),
    {
      onError,
      refetchInterval: 5000,
    },
  );
}
