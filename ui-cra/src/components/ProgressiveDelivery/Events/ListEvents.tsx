import React from 'react';
import { useQuery } from 'react-query';
import {
  ClustersService,
  ListEventsRequest,
  ListEventsResponse,
} from '../../../cluster-services/cluster_services.pb';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Alert } from '@material-ui/lab';
import { EventsTable } from './EventsTable';

const ListEvents = (listEventsPayload: ListEventsRequest) => {
  const { error, data, isLoading } = useQuery<ListEventsResponse, Error>(
    'events',
    () => ClustersService.ListEvents(listEventsPayload),
  );
  return (
    <>
      {isLoading && <LoadingPage />}
      {error && <Alert severity="error">{error.message}</Alert>}
      {data?.events && <EventsTable events={data.events} />}
    </>
  );
};

export default ListEvents;
