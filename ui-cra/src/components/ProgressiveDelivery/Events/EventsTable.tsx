import { filterConfig, theme, FilterableTable } from '@weaveworks/weave-gitops';
import moment from 'moment';
import { ThemeProvider } from 'styled-components';
import { Event } from '../../../cluster-services/cluster_services.pb';
import { usePolicyStyle } from '../../Policies/PolicyStyles';
import { TableWrapper } from '../CanaryStyles';

export const EventsTable = ({ events }: { events: Event[] }) => {
  const classes = usePolicyStyle();

  const initialFilterState = {
    ...filterConfig(events, 'component'),
  };

  return (
    <div className={classes.root}>
      <ThemeProvider theme={theme}>
        {events.length > 0 ? (
          <TableWrapper id="events-list">
            <FilterableTable
              key={events?.length}
              filters={initialFilterState}
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
                    <span className={classes.messageWrape}>{e.message}</span>
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
        ) : (
          <p>No data to display</p>
        )}
      </ThemeProvider>
    </div>
  );
};
