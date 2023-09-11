import { DataTable } from '@weaveworks/weave-gitops';
import moment from 'moment';
import { Event } from '../../../../cluster-services/cluster_services.pb';
import { TableWrapper } from '../../../Shared';

export const EventsTable = ({ events }: { events: Event[] }) => {
  return (
    <TableWrapper id="events-list">
      <DataTable
        rows={events}
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
            label: 'Age',
            value: (e: Event) => moment(e.timestamp).fromNow() || '--',
            defaultSort: true,
            sortValue: (e: Event) => {
              const t = e.timestamp && new Date(e.timestamp);
              // Invert the timestamp so the default sort shows the most recent first
              return e.timestamp ? (Number(t) * -1).toString() : '';
            },
          },
        ]}
      />
    </TableWrapper>
  );
};
