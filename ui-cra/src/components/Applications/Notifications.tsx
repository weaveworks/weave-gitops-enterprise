import { FC } from 'react';
import {
  NotificationsTable,
  Page,
  useListProviders,
} from '@weaveworks/weave-gitops';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { Provider } from '@weaveworks/weave-gitops/ui/lib/objects';

const WGNotifications: FC = () => {
  const { data, isLoading, error } = useListProviders();

  return (
    <Page
      path={[
        {
          label: 'Notifications',
          url: '/notifications',
        },
      ]}
    >
      <ContentWrapper
        loading={isLoading}
        errors={error ? [{ message: error?.message }] : []}
      >
        <NotificationsTable rows={data?.objects as Provider[]} />
      </ContentWrapper>
    </Page>
  );
};

export default WGNotifications;
