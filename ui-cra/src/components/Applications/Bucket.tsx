import { FC } from 'react';
import { BucketDetail, Kind, useGetObject } from '@weaveworks/weave-gitops';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { Bucket } from '@weaveworks/weave-gitops/ui/lib/objects';
import { EditButton } from '../EditButton';

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
    <PageTemplate
      documentTitle="Bucket"
      path={[
        {
          label: 'Applications',
          url: '/applications',
        },
        {
          label: 'Sources',
          url: '/sources',
        },
        {
          label: `${props.name}`,
        },
      ]}
    >
      <ContentWrapper
        loading={isLoading}
        errors={
          error ? [{ clusterName, namespace, message: error?.message }] : []
        }
      >
        <BucketDetail
          bucket={bucket}
          customActions={[<EditButton resource={bucket} />]}
          {...props}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsBucket;
