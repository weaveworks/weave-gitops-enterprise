import { FC } from 'react';
import { BucketDetail, useListSources } from '@weaveworks/weave-gitops';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { useApplicationsCount, useSourcesCount } from './utils';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

const WGApplicationsBucket: FC<Props> = props => {
  const applicationsCount = useApplicationsCount();
  const { isLoading } = useListSources();
  const sourcesCount = useSourcesCount();

  return (
    <PageTemplate documentTitle="WeGO · Bucket">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
          {
            label: 'Sources',
            url: '/sources',
            count: sourcesCount,
          },
          {
            label: `${props.name}`,
          },
        ]}
      />
      <ContentWrapper loading={isLoading}>
        <BucketDetail {...props} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsBucket;
