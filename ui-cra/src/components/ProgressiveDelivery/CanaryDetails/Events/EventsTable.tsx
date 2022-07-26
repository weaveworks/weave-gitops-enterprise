import moment from 'moment';
import { Event } from '../../../../cluster-services/cluster_services.pb';
import { CustomDataTable, TableWrapper } from '../../CanaryStyles';

export const EventsTable = ({ events }: { events: Event[] }) => {
  return (
    <TableWrapper id="events-list">
      <CustomDataTable
        rows={events}
        defaultSort={3}
        fields={[
          {
            label: 'Reason',
            value: 'reason',
            textSearchable: true,
          },
          {
            label: 'Message',
            value: (e: Event) => (
              <span
                style={{
                  whiteSpace: 'normal',
                }}
              >
                {e.message}
              </span>
            ),
            maxWidth: 600,
          },
          {
            label: 'From',
            value: 'component',
          },
          {
            label: 'Last Updated',
            value: (e: Event) => moment(e.timestamp).fromNow() || '--',
            sortValue: (e: Event) => e.timestamp,
          },
        ]}
      />
    </TableWrapper>
  );
};
