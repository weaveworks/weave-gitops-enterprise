import { FC } from 'react';
import {
  BucketDetail,
  Kind,
  Page,
  useGetObject,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { Bucket } from '@weaveworks/weave-gitops/ui/lib/objects';
import { EditButton } from '../Templates/Edit/EditButton';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

const WGApplicationsBucket: FC<Props> = props => {
  const { name, namespace, clusterName } = props;
  const {
    data: bucket,
    isLoading,
    error,
  } = useGetObject<Bucket>(name, namespace, Kind.Bucket, clusterName);

  return (
    <Page
      loading={isLoading}
      path={[
        {
          label: 'Sources',
          url: V2Routes.Sources,
        },
        {
          label: `${props.name}`,
        },
      ]}
    >
      <NotificationsWrapper
        errors={
          error ? [{ clusterName, namespace, message: error?.message }] : []
        }
      >
        <BucketDetail
          bucket={bucket}
          customActions={[<EditButton resource={bucket} />]}
          {...props}
        />
      </NotificationsWrapper>
    </Page>
  );
};

export default WGApplicationsBucket;
