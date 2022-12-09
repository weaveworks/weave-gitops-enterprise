import { FC } from 'react';
import {
  Kind,
  ProviderDetail,
  useGetObject,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
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
    <PageTemplate
      documentTitle="Notification Providers"
      path={[
        {
          label: 'Notification Providers',
          url: V2Routes.Provider,
        },
      ]}
    >
      <ContentWrapper
        loading={isLoading}
        errors={error ? [{ message: error?.message }] : []}
      >
        <ProviderDetail provider={data} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGNotificationsProvider;
