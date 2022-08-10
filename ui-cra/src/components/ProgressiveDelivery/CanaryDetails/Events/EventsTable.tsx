import { SortType } from '@weaveworks/weave-gitops';
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
            sortType: SortType.number,
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
