import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { ListEventsRequest } from '../../../../cluster-services/cluster_services.pb';
import { useListEvents } from '../../../../contexts/ProgressiveDelivery';
import { EventsTable } from './EventsTable';

type Props = ListEventsRequest;

const ListEvents = (props: Props) => {
  const { error, data, isLoading } = useListEvents(props);
  return (
    <>
      {isLoading && <LoadingPage />}
      {error && <Alert severity="error">{error.message}</Alert>}
      {data?.events && <EventsTable events={data.events} />}
    </>
  );
};

export default ListEvents;
