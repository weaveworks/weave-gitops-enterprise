import { FC } from 'react';
import {
  NotificationsTable,
  Page,
  useListProviders,
} from '@weaveworks/weave-gitops';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { Provider } from '@weaveworks/weave-gitops/ui/lib/objects';

const WGNotifications: FC = () => {
  const { data, isLoading, error } = useListProviders();

  return (
    <Page
      loading={isLoading}
      path={[
        {
          label: 'Notifications',
          url: '/notifications',
        },
      ]}
    >
      <NotificationsWrapper errors={error ? [{ message: error?.message }] : []}>
        <NotificationsTable rows={data?.objects as Provider[]} />
      </NotificationsWrapper>
    </Page>
  );
};

export default WGNotifications;
