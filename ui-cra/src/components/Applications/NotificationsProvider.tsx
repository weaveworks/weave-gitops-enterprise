import { FC } from 'react';
import {
  Kind,
  Page,
  ProviderDetail,
  useGetObject,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { Provider } from '@weaveworks/weave-gitops/ui/lib/objects';

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
      error={error ? [{ message: error?.message }] : []}
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
      <ProviderDetail provider={data} />
    </Page>
  );
};

export default WGNotificationsProvider;
