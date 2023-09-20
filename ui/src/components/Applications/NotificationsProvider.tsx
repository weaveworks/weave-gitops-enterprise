import { Provider } from '@weaveworks/weave-gitops/ui/lib/objects';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import {
  Kind,
  ProviderDetail,
  useGetObject,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { FC } from 'react';

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

const WGNotificationsProvider: FC<Props> = ({
  name,
  namespace,
  clusterName,
}) => {
  const { data, isLoading, error } = useGetObject<Provider>(
    name,
    namespace,
    Kind.Provider,
    clusterName,
  );

  return (
    <Page
      loading={isLoading}
      path={[
        {
          label: 'Notifications',
          url: V2Routes.Notifications,
        },
        {
          label: name,
        },
      ]}
    >
      <NotificationsWrapper errors={error ? [{ message: error?.message }] : []}>
        <ProviderDetail provider={data} />
      </NotificationsWrapper>
    </Page>
  );
};

export default WGNotificationsProvider;
