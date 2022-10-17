import { FC } from 'react';
import { BucketDetail, Kind, useGetObject, V2Routes } from '@weaveworks/weave-gitops';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { Bucket } from '@weaveworks/weave-gitops/ui/lib/objects';
import { Routes } from '../../utils/nav';

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
          url: Routes.Applications,
        },
        {
          label: 'Sources',
          url: V2Routes.Sources,
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
        <BucketDetail bucket={bucket} {...props} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsBucket;
