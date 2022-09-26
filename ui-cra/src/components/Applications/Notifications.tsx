import { FC } from 'react';
import { NotificationsTable, useListProviders } from '@weaveworks/weave-gitops';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { Provider } from '@weaveworks/weave-gitops/ui/lib/objects';

const WGNotifications: FC = () => {
  const { data, isLoading, error } = useListProviders();

  return (
    <PageTemplate
      documentTitle="WeGO · Notifications"
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
    </PageTemplate>
  );
};

export default WGNotifications;
