import { ListEventsRequest } from '../../../../cluster-services/cluster_services.pb';
import { useListEvents } from '../../../../contexts/ProgressiveDelivery';
import LoadingWrapper from '../../../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';
import { EventsTable } from './EventsTable';

type Props = ListEventsRequest;

const ListEvents = (props: Props) => {
  const { error, data, isLoading } = useListEvents(props);
  return (
    <>
      <LoadingWrapper
        loading={isLoading}
        errorMessage={error?.message || ''}
      >
        <EventsTable events={data?.events || []} />
      </LoadingWrapper>
    </>
  );
};

export default ListEvents;
