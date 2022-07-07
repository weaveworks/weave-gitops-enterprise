import {
  ListCanariesResponse,
  ProgressiveDeliveryService,
} from '@weaveworks/progressive-delivery';
import _ from 'lodash';
import React from 'react';
import { useQuery } from 'react-query';

interface Props {
  api: typeof ProgressiveDeliveryService;
  children: any;
}

export const ProgressiveDeliveryContext = React.createContext<
  typeof ProgressiveDeliveryService
>(null as any);

export const ProgressiveDeliveryProvider = ({ api, children }: Props) => (
  <ProgressiveDeliveryContext.Provider value={api}>
    {children}
  </ProgressiveDeliveryContext.Provider>
);

export const useProgressiveDelivery = () =>
  React.useContext(ProgressiveDeliveryContext);

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
  );
};
