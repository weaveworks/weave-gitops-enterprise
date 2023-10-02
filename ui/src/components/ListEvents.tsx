import {
  DataTable,
  Icon,
  IconType,
  RequestStateHandler,
  Text,
  Timestamp,
} from '@weaveworks/weave-gitops';

import { RequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import {
  Event,
  ListEventsRequest,
} from '../cluster-services/cluster_services.pb';
import { useListEvents } from '../contexts/ProgressiveDelivery';
import styled from 'styled-components';

const Reason = styled.h2`
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px !important;
`;

const ListEvents = (props: ListEventsRequest) => {
  const { error, data, isLoading } = useListEvents(props);
  return (
    <RequestStateHandler loading={isLoading} error={error as RequestError}>
      <DataTable
        fields={[
          {
            label: 'Reason',
            labelRenderer: () => {
              return (
                <Reason title="It refers to what triggered the event. It can be different according to each component.">
                  Reason
                  <Icon
                    size="base"
                    type={IconType.InfoIcon}
                    color="neutral30"
                  />
                </Reason>
              );
            },
            value: (e: Event) => <Text capitalize>{e?.reason}</Text>,
            sortValue: (e: Event) => e?.reason,
          },
          { label: 'Message', value: 'message', maxWidth: 600 },
          { label: 'From', value: 'component' },
          {
            label: 'Last Updated',
            value: (e: Event) => <Timestamp time={e?.timestamp || ''} />,
            sortValue: (e: Event) => -Date.parse(e?.timestamp || ''),
            defaultSort: true,
            secondarySort: true,
          },
        ]}
        rows={data?.events}
      />
    </RequestStateHandler>
  );
};

export default ListEvents;
