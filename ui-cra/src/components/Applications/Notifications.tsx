import { FC } from 'react';
import {
  NotificationsTable,
  Page,
  useListProviders,
} from '@weaveworks/weave-gitops';
import { Provider } from '@weaveworks/weave-gitops/ui/lib/objects';

const WGNotifications: FC = () => {
  const { data, isLoading, error } = useListProviders();

  return (
    <Page
      loading={isLoading}
      error={error ? [{ message: error?.message }] : []}
      path={[
        {
          label: 'Notifications',
          url: '/notifications',
        },
      ]}
    >
      <NotificationsTable rows={data?.objects as Provider[]} />
    </Page>
  );
};

export default WGNotifications;
