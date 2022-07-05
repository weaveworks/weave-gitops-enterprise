import { ListEventsRequest } from '../../../cluster-services/cluster_services.pb';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Alert } from '@material-ui/lab';
import { EventsTable } from './EventsTable';
import { useListEvents } from '../../../hooks/progressiveDelivery';

const ListEvents = (listEventsPayload: ListEventsRequest) => {
  const { error, data, isLoading } = useListEvents(listEventsPayload);
  return (
    <>
      {isLoading && <LoadingPage />}
      {error && <Alert severity="error">{error.message}</Alert>}
      {data?.events && <EventsTable events={data.events} />}
    </>
  );
};

export default ListEvents;
