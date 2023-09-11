import { FC } from 'react';
import { NotificationsTable, useListProviders } from '@weaveworks/weave-gitops';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { Provider } from '@weaveworks/weave-gitops/ui/lib/objects';
import { Page } from '../Layout/App';

const WGNotifications: FC = () => {
  const { data, isLoading, error } = useListProviders();

  return (
    <Page
      loading={isLoading}
      path={[
        {
          label: 'Notifications',
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
